package repl

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"

	prompt "github.com/c-bata/go-prompt"
	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/interpreter"
	"github.com/influxdata/ifql/parser"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/control"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/semantic"
	"github.com/pkg/errors"
)

type REPL struct {
	scope        *interpreter.Scope
	declarations semantic.DeclarationScope
	c            *control.Controller
	d            query.Domain

	cancelMu   sync.Mutex
	cancelFunc context.CancelFunc
}

func addBuiltIn(script string, d query.Domain, scope *interpreter.Scope, declarations semantic.DeclarationScope) error {
	astProg, err := parser.NewAST(script)
	if err != nil {
		return errors.Wrap(err, "failed to parse builtin")
	}
	semProg, err := semantic.New(astProg, declarations)
	if err != nil {
		return errors.Wrap(err, "failed to create semantic graph for builtin")
	}

	if err := interpreter.Eval(semProg, scope, d); err != nil {
		return errors.Wrap(err, "failed to evaluate builtin")
	}
	return nil
}

func New(c *control.Controller) *REPL {
	scope, declarations := query.BuiltIns()
	d := query.NewDomain()
	addBuiltIn("run = () => yield(table:_)", d, scope, declarations)
	return &REPL{
		scope:        scope,
		declarations: declarations,
		c:            c,
		d:            d,
	}
}

func (r *REPL) Run() {
	p := prompt.New(
		r.input,
		r.completer,
		prompt.OptionPrefix("> "),
		prompt.OptionTitle("ifql"),
	)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)
	go func() {
		for range sigs {
			r.cancel()
		}
	}()
	p.Run()
}

func (r *REPL) cancel() {
	r.cancelMu.Lock()
	defer r.cancelMu.Unlock()
	if r.cancelFunc != nil {
		r.cancelFunc()
		r.cancelFunc = nil
	}
}

func (r *REPL) setCancel(cf context.CancelFunc) {
	r.cancelMu.Lock()
	defer r.cancelMu.Unlock()
	r.cancelFunc = cf
}
func (r *REPL) clearCancel() {
	r.setCancel(nil)
}

func (r *REPL) completer(d prompt.Document) []prompt.Suggest {
	names := r.scope.Names()
	sort.Strings(names)

	s := make([]prompt.Suggest, 0, len(names))
	for _, n := range names {
		if n == "_" || !strings.HasPrefix(n, "_") {
			s = append(s, prompt.Suggest{Text: n})
		}
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func (r *REPL) Input(t string) error {
	_, err := r.executeLine(t, false)
	return err
}

// input processes a line of input and prints the result.
func (r *REPL) input(t string) {
	v, err := r.executeLine(t, true)
	if err != nil {
		fmt.Println("Error:", err)
	} else if v != nil {
		fmt.Println(v)
	}
}

// executeLine processes a line of input.
// If the input evaluates to a valid value, that value is returned.
func (r *REPL) executeLine(t string, expectYield bool) (interpreter.Value, error) {
	if t == "" {
		return nil, nil
	}
	astProg, err := parser.NewAST(t)
	if err != nil {
		return nil, err
	}

	semProg, err := semantic.New(astProg, r.declarations)
	if err != nil {
		return nil, err
	}

	if err := interpreter.Eval(semProg, r.scope, r.d); err != nil {
		return nil, err
	}

	v := r.scope.Return()

	// Check for yield and execute query
	if v.Type() == query.TableObjectType {
		t := v.(query.TableObject)
		if !expectYield || (expectYield && t.Kind() == functions.YieldKind) {
			spec := t.ToSpec()
			// Do query
			return nil, r.doQuery(spec)
		}
	}

	r.scope.Set("_", v)

	// Print value
	if v.Type() != semantic.Invalid {
		return v, nil
	}
	return nil, nil
}

func (r *REPL) doQuery(spec *query.Spec) error {
	// Setup cancel context
	ctx, cancelFunc := context.WithCancel(context.Background())
	r.setCancel(cancelFunc)
	defer cancelFunc()
	defer r.clearCancel()

	q, err := r.c.Query(ctx, spec)
	if err != nil {
		return err
	}
	defer q.Done()

	results, ok := <-q.Ready
	if !ok {
		err := q.Err()
		return err
	}

	names := make([]string, 0, len(results))
	for name := range results {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		r := results[name]
		blocks := r.Blocks()
		fmt.Println("Result:", name)
		err := blocks.Do(func(b execute.Block) error {
			execute.NewFormatter(b, nil).WriteTo(os.Stdout)
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
