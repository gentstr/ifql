package functions_test

import (
	"testing"

	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/execute/executetest"
	"github.com/influxdata/ifql/query/plan"
	"github.com/influxdata/ifql/query/plan/plantest"
	"github.com/influxdata/ifql/query/querytest"
)

func TestFirstOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"first","kind":"first","spec":{"useRowTime":true}}`)
	op := &query.Operation{
		ID: "first",
		Spec: &functions.FirstOpSpec{
			UseRowTime: true,
		},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestFirst_Process(t *testing.T) {
	testCases := []struct {
		name string
		data *executetest.Block
		want [][]int
	}{
		{
			name: "first",
			data: &executetest.Block{
				ColMeta: []execute.ColMeta{
					{Label: "time", Type: execute.TTime},
					{Label: "value", Type: execute.TFloat},
					{Label: "t1", Type: execute.TString, IsTag: true, IsCommon: true},
					{Label: "t2", Type: execute.TString, IsTag: true, IsCommon: false},
				},
				Data: [][]interface{}{
					{execute.Time(0), 0.0, "a", "y"},
					{execute.Time(10), 5.0, "a", "x"},
					{execute.Time(20), 9.0, "a", "y"},
					{execute.Time(30), 4.0, "a", "x"},
					{execute.Time(40), 6.0, "a", "y"},
					{execute.Time(50), 8.0, "a", "x"},
					{execute.Time(60), 1.0, "a", "y"},
					{execute.Time(70), 2.0, "a", "x"},
					{execute.Time(80), 3.0, "a", "y"},
					{execute.Time(90), 7.0, "a", "x"},
				},
			},
			want: [][]int{{0}, nil, nil, nil, nil, nil, nil, nil, nil, nil},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			executetest.IndexSelectorFuncTestHelper(
				t,
				new(functions.FirstSelector),
				tc.data,
				tc.want,
			)
		})
	}
}

func BenchmarkFirst(b *testing.B) {
	executetest.IndexSelectorFuncBenchmarkHelper(b, new(functions.FirstSelector), NormalBlock)
}

func TestFirst_PushDown_Single(t *testing.T) {
	lp := &plan.LogicalPlanSpec{
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			plan.ProcedureIDFromOperationID("from"): {
				ID: plan.ProcedureIDFromOperationID("from"),
				Spec: &functions.FromProcedureSpec{
					Database: "mydb",
				},
				Parents:  nil,
				Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("first")},
			},
			plan.ProcedureIDFromOperationID("first"): {
				ID:   plan.ProcedureIDFromOperationID("first"),
				Spec: &functions.FirstProcedureSpec{},
				Parents: []plan.ProcedureID{
					(plan.ProcedureIDFromOperationID("from")),
				},
				Children: nil,
			},
		},
		Order: []plan.ProcedureID{
			plan.ProcedureIDFromOperationID("from"),
			plan.ProcedureIDFromOperationID("first"),
		},
	}

	want := &plan.PlanSpec{
		Bounds: plan.BoundsSpec{
			Start: query.MinTime,
			Stop:  query.Now,
		},
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			plan.ProcedureIDFromOperationID("from"): {
				ID: plan.ProcedureIDFromOperationID("from"),
				Spec: &functions.FromProcedureSpec{
					Database:  "mydb",
					BoundsSet: true,
					Bounds: plan.BoundsSpec{
						Start: query.MinTime,
						Stop:  query.Now,
					},
					LimitSet:      true,
					PointsLimit:   1,
					DescendingSet: true,
					Descending:    false,
				},
				Children: []plan.ProcedureID{},
			},
		},
		Results: []plan.ProcedureID{
			(plan.ProcedureIDFromOperationID("from")),
		},
		Order: []plan.ProcedureID{
			plan.ProcedureIDFromOperationID("from"),
		},
	}

	plantest.PhysicalPlanTestHelper(t, lp, want)
}

func TestFirst_PushDown_Branch(t *testing.T) {
	lp := &plan.LogicalPlanSpec{
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			plan.ProcedureIDFromOperationID("from"): {
				ID: plan.ProcedureIDFromOperationID("from"),
				Spec: &functions.FromProcedureSpec{
					Database: "mydb",
				},
				Parents: nil,
				Children: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("first"),
					plan.ProcedureIDFromOperationID("last"),
				},
			},
			plan.ProcedureIDFromOperationID("first"): {
				ID:       plan.ProcedureIDFromOperationID("first"),
				Spec:     &functions.FirstProcedureSpec{},
				Parents:  []plan.ProcedureID{plan.ProcedureIDFromOperationID("from")},
				Children: nil,
			},
			plan.ProcedureIDFromOperationID("last"): {
				ID:       plan.ProcedureIDFromOperationID("last"),
				Spec:     &functions.LastProcedureSpec{},
				Parents:  []plan.ProcedureID{plan.ProcedureIDFromOperationID("from")},
				Children: nil,
			},
		},
		Order: []plan.ProcedureID{
			plan.ProcedureIDFromOperationID("from"),
			plan.ProcedureIDFromOperationID("last"),
			plan.ProcedureIDFromOperationID("first"), // first is last so it will be duplicated
		},
	}

	fromID := plan.ProcedureIDFromOperationID("from")
	fromIDDup := plan.ProcedureIDForDuplicate(fromID)
	want := &plan.PlanSpec{
		Bounds: plan.BoundsSpec{
			Start: query.MinTime,
			Stop:  query.Now,
		},
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			fromID: {
				ID: fromID,
				Spec: &functions.FromProcedureSpec{
					Database:  "mydb",
					BoundsSet: true,
					Bounds: plan.BoundsSpec{
						Start: query.MinTime,
						Stop:  query.Now,
					},
					LimitSet:      true,
					PointsLimit:   1,
					DescendingSet: true,
					Descending:    true, // last
				},
				Children: []plan.ProcedureID{},
			},
			fromIDDup: {
				ID: fromIDDup,
				Spec: &functions.FromProcedureSpec{
					Database:  "mydb",
					BoundsSet: true,
					Bounds: plan.BoundsSpec{
						Start: query.MinTime,
						Stop:  query.Now,
					},
					LimitSet:      true,
					PointsLimit:   1,
					DescendingSet: true,
					Descending:    false, // fist
				},
				Parents:  []plan.ProcedureID{},
				Children: []plan.ProcedureID{},
			},
		},
		Results: []plan.ProcedureID{
			fromID,
			fromIDDup,
		},
		Order: []plan.ProcedureID{
			fromID,
			fromIDDup,
		},
	}

	plantest.PhysicalPlanTestHelper(t, lp, want)
}
