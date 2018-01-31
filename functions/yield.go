package functions

import (
	"fmt"

	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/plan"
	"github.com/influxdata/ifql/semantic"
)

const YieldKind = "yield"

type YieldOpSpec struct {
	Name string `json:"name"`
}

var yieldSignature = semantic.FunctionSignature{
	Params: map[string]semantic.Type{
		query.TableParameter: query.TableObjectType,
		"name":               semantic.String,
	},
	ReturnType:   semantic.Nil,
	PipeArgument: query.TableParameter,
}

func init() {
	query.RegisterFunction(YieldKind, createYieldOpSpec, yieldSignature)
	query.RegisterOpSpec(YieldKind, newYieldOp)
	plan.RegisterProcedureSpec(YieldKind, newYieldProcedure, YieldKind)
}

func createYieldOpSpec(args query.Arguments, a *query.Administration) (query.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(YieldOpSpec)

	name, err := args.GetRequiredString("name")
	if err != nil {
		return nil, err
	}
	spec.Name = name

	return spec, nil
}

func newYieldOp() query.OperationSpec {
	return new(YieldOpSpec)
}

func (s *YieldOpSpec) Kind() query.OperationKind {
	return YieldKind
}

type YieldProcedureSpec struct {
	Name string `json:"name"`
}

func newYieldProcedure(qs query.OperationSpec, _ plan.Administration) (plan.ProcedureSpec, error) {
	if spec, ok := qs.(*YieldOpSpec); ok {
		return &YieldProcedureSpec{Name: spec.Name}, nil
	}

	return nil, fmt.Errorf("invalid spec type %T", qs)
}

func (s *YieldProcedureSpec) Kind() plan.ProcedureKind {
	return YieldKind
}

func (s *YieldProcedureSpec) Copy() plan.ProcedureSpec {
	return &YieldProcedureSpec{Name: s.Name}
}

func (s *YieldProcedureSpec) YieldName() string {
	return s.Name
}
