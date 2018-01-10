package functions_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/influxdata/ifql/ast"
	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/execute/executetest"
	"github.com/influxdata/ifql/query/plan"
	"github.com/influxdata/ifql/query/plan/plantest"
	"github.com/influxdata/ifql/query/querytest"
)

func TestFilter_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "from with database filter and range",
			Raw:  `from(db:"mydb").filter(fn: (r) => r["t1"]=="val1" and r["t2"]=="val2").range(start:-4h, stop:-2h).count()`,
			Want: &query.Spec{
				Operations: []*query.Operation{
					{
						ID: "from0",
						Spec: &functions.FromOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "filter1",
						Spec: &functions.FilterOpSpec{
							Fn: &ast.ArrowFunctionExpression{
								Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
								Body: &ast.LogicalExpression{
									Operator: ast.AndOperator,
									Left: &ast.BinaryExpression{
										Operator: ast.EqualOperator,
										Left: &ast.MemberExpression{
											Object: &ast.Identifier{
												Name: "r",
											},
											Property: &ast.StringLiteral{Value: "t1"},
										},
										Right: &ast.StringLiteral{Value: "val1"},
									},
									Right: &ast.BinaryExpression{
										Operator: ast.EqualOperator,
										Left: &ast.MemberExpression{
											Object: &ast.Identifier{
												Name: "r",
											},
											Property: &ast.StringLiteral{Value: "t2"},
										},
										Right: &ast.StringLiteral{Value: "val2"},
									},
								},
							},
						},
					},
					{
						ID: "range2",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: query.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
						},
					},
					{
						ID:   "count3",
						Spec: &functions.CountOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "from0", Child: "filter1"},
					{Parent: "filter1", Child: "range2"},
					{Parent: "range2", Child: "count3"},
				},
			},
		},
		{
			Name: "from with database filter (and with or) and range",
			Raw: `from(db:"mydb")
						.filter(fn: (r) =>
								(
									(r["t1"]=="val1")
									and
									(r["t2"]=="val2")
								)
								or
								(r["t3"]=="val3")
							)
						.range(start:-4h, stop:-2h)
						.count()`,
			Want: &query.Spec{
				Operations: []*query.Operation{
					{
						ID: "from0",
						Spec: &functions.FromOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "filter1",
						Spec: &functions.FilterOpSpec{
							Fn: &ast.ArrowFunctionExpression{
								Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
								Body: &ast.LogicalExpression{
									Operator: ast.OrOperator,
									Left: &ast.LogicalExpression{
										Operator: ast.AndOperator,
										Left: &ast.BinaryExpression{
											Operator: ast.EqualOperator,
											Left: &ast.MemberExpression{
												Object: &ast.Identifier{
													Name: "r",
												},
												Property: &ast.StringLiteral{Value: "t1"},
											},
											Right: &ast.StringLiteral{Value: "val1"},
										},
										Right: &ast.BinaryExpression{
											Operator: ast.EqualOperator,
											Left: &ast.MemberExpression{
												Object: &ast.Identifier{
													Name: "r",
												},
												Property: &ast.StringLiteral{Value: "t2"},
											},
											Right: &ast.StringLiteral{Value: "val2"},
										},
									},
									Right: &ast.BinaryExpression{
										Operator: ast.EqualOperator,
										Left: &ast.MemberExpression{
											Object: &ast.Identifier{
												Name: "r",
											},
											Property: &ast.StringLiteral{Value: "t3"},
										},
										Right: &ast.StringLiteral{Value: "val3"},
									},
								},
							},
						},
					},
					{
						ID: "range2",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: query.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
						},
					},
					{
						ID:   "count3",
						Spec: &functions.CountOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "from0", Child: "filter1"},
					{Parent: "filter1", Child: "range2"},
					{Parent: "range2", Child: "count3"},
				},
			},
		},
		{
			Name: "from with database filter including fields",
			Raw: `from(db:"mydb")
						.filter(fn: (r) =>
							(r["t1"] =="val1")
							and
							(r["_field"] == 10)
						)
						.range(start:-4h, stop:-2h)
						.count()`,
			Want: &query.Spec{
				Operations: []*query.Operation{
					{
						ID: "from0",
						Spec: &functions.FromOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "filter1",
						Spec: &functions.FilterOpSpec{
							Fn: &ast.ArrowFunctionExpression{
								Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
								Body: &ast.LogicalExpression{
									Operator: ast.AndOperator,
									Left: &ast.BinaryExpression{
										Operator: ast.EqualOperator,
										Left: &ast.MemberExpression{
											Object: &ast.Identifier{
												Name: "r",
											},
											Property: &ast.StringLiteral{Value: "t1"},
										},
										Right: &ast.StringLiteral{Value: "val1"},
									},
									Right: &ast.BinaryExpression{
										Operator: ast.EqualOperator,
										Left: &ast.MemberExpression{
											Object: &ast.Identifier{
												Name: "r",
											},
											Property: &ast.StringLiteral{Value: "_field"},
										},
										Right: &ast.IntegerLiteral{Value: 10},
									},
								},
							},
						},
					},
					{
						ID: "range2",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: query.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
						},
					},
					{
						ID:   "count3",
						Spec: &functions.CountOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "from0", Child: "filter1"},
					{Parent: "filter1", Child: "range2"},
					{Parent: "range2", Child: "count3"},
				},
			},
		},
		{
			Name: "from with database filter with no parens including fields",
			Raw: `from(db:"mydb")
						.filter(fn: (r) =>
							r["t1"]=="val1"
							and
							r["_field"] == 10
						)
						.range(start:-4h, stop:-2h)
						.count()`,
			Want: &query.Spec{
				Operations: []*query.Operation{
					{
						ID: "from0",
						Spec: &functions.FromOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "filter1",
						Spec: &functions.FilterOpSpec{
							Fn: &ast.ArrowFunctionExpression{
								Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
								Body: &ast.LogicalExpression{
									Operator: ast.AndOperator,
									Left: &ast.BinaryExpression{
										Operator: ast.EqualOperator,
										Left: &ast.MemberExpression{
											Object: &ast.Identifier{
												Name: "r",
											},
											Property: &ast.StringLiteral{Value: "t1"},
										},
										Right: &ast.StringLiteral{Value: "val1"},
									},
									Right: &ast.BinaryExpression{
										Operator: ast.EqualOperator,
										Left: &ast.MemberExpression{
											Object: &ast.Identifier{
												Name: "r",
											},
											Property: &ast.StringLiteral{Value: "_field"},
										},
										Right: &ast.IntegerLiteral{Value: 10},
									},
								},
							},
						},
					},
					{
						ID: "range2",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: query.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
						},
					},
					{
						ID:   "count3",
						Spec: &functions.CountOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "from0", Child: "filter1"},
					{Parent: "filter1", Child: "range2"},
					{Parent: "range2", Child: "count3"},
				},
			},
		},
		{
			Name: "from with database filter with no parens including regex and field",
			Raw: `from(db:"mydb")
						.filter(fn: (r) =>
							r["t1"]==/val1/
							and
							r["_field"] == 10.5
						)
						.range(start:-4h, stop:-2h)
						.count()`,
			Want: &query.Spec{
				Operations: []*query.Operation{
					{
						ID: "from0",
						Spec: &functions.FromOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "filter1",
						Spec: &functions.FilterOpSpec{
							Fn: &ast.ArrowFunctionExpression{
								Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
								Body: &ast.LogicalExpression{
									Operator: ast.AndOperator,
									Left: &ast.BinaryExpression{
										Operator: ast.EqualOperator,
										Left: &ast.MemberExpression{
											Object: &ast.Identifier{
												Name: "r",
											},
											Property: &ast.StringLiteral{Value: "t1"},
										},
										Right: &ast.RegexpLiteral{Value: regexp.MustCompile("val1")},
									},
									Right: &ast.BinaryExpression{
										Operator: ast.EqualOperator,
										Left: &ast.MemberExpression{
											Object: &ast.Identifier{
												Name: "r",
											},
											Property: &ast.StringLiteral{Value: "_field"},
										},
										Right: &ast.FloatLiteral{Value: 10.5},
									},
								},
							},
						},
					},
					{
						ID: "range2",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: query.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
						},
					},
					{
						ID:   "count3",
						Spec: &functions.CountOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "from0", Child: "filter1"},
					{Parent: "filter1", Child: "range2"},
					{Parent: "range2", Child: "count3"},
				},
			},
		},
		{
			Name: "from with database regex with escape",
			Raw: `from(db:"mydb")
						.filter(fn: (r) => 
							r["t1"]==/va\/l1/
						)`,
			Want: &query.Spec{
				Operations: []*query.Operation{
					{
						ID: "from0",
						Spec: &functions.FromOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "filter1",
						Spec: &functions.FilterOpSpec{
							Fn: &ast.ArrowFunctionExpression{
								Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
								Body: &ast.BinaryExpression{
									Operator: ast.EqualOperator,
									Left: &ast.MemberExpression{
										Object: &ast.Identifier{
											Name: "r",
										},
										Property: &ast.StringLiteral{Value: "t1"},
									},
									Right: &ast.RegexpLiteral{Value: regexp.MustCompile(`va/l1`)},
								},
							},
						},
					},
				},
				Edges: []query.Edge{
					{Parent: "from0", Child: "filter1"},
				},
			},
		},
		{
			Name: "from with database with two regex",
			Raw: `from(db:"mydb")
						.filter(fn: (r) => 
							r["t1"]==/va\/l1/
							and
							r["t2"] != /val2/
						)`,
			Want: &query.Spec{
				Operations: []*query.Operation{
					{
						ID: "from0",
						Spec: &functions.FromOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "filter1",
						Spec: &functions.FilterOpSpec{
							Fn: &ast.ArrowFunctionExpression{
								Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
								Body: &ast.LogicalExpression{
									Operator: ast.AndOperator,
									Left: &ast.BinaryExpression{
										Operator: ast.EqualOperator,
										Left: &ast.MemberExpression{
											Object: &ast.Identifier{
												Name: "r",
											},
											Property: &ast.StringLiteral{Value: "t1"},
										},
										Right: &ast.RegexpLiteral{Value: regexp.MustCompile(`va/l1`)},
									},
									Right: &ast.BinaryExpression{
										Operator: ast.NotEqualOperator,
										Left: &ast.MemberExpression{
											Object: &ast.Identifier{
												Name: "r",
											},
											Property: &ast.StringLiteral{Value: "t2"},
										},
										Right: &ast.RegexpLiteral{Value: regexp.MustCompile(`val2`)},
									},
								},
							},
						},
					},
				},
				Edges: []query.Edge{
					{Parent: "from0", Child: "filter1"},
				},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			querytest.NewQueryTestHelper(t, tc)
		})
	}
}
func TestFilterOperation_Marshaling(t *testing.T) {
	data := []byte(`{
		"id":"filter",
		"kind":"filter",
		"spec":{
			"fn":{
				"type": "ArrowFunctionExpression",
				"params": [{"type":"Property","key":{"type":"Identifier","name":"r"}}],
				"body":{
					"type":"BinaryExpression",
					"operator": "!=",
					"left":{
						"type":"MemberExpression",
						"object": {
							"type": "Identifier",
							"name":"r"
						},
						"property": {"type":"StringLiteral","value":"_measurement"}
					},
					"right":{
						"type":"StringLiteral",
						"value":"mem"
					}
				}
			}
		}
	}`)
	op := &query.Operation{
		ID: "filter",
		Spec: &functions.FilterOpSpec{
			Fn: &ast.ArrowFunctionExpression{
				Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
				Body: &ast.BinaryExpression{
					Operator: ast.NotEqualOperator,
					Left: &ast.MemberExpression{
						Object: &ast.Identifier{
							Name: "r",
						},
						Property: &ast.StringLiteral{Value: "_measurement"},
					},
					Right: &ast.StringLiteral{Value: "mem"},
				},
			},
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestFilter_Process(t *testing.T) {
	testCases := []struct {
		name string
		spec *functions.FilterProcedureSpec
		data []execute.Block
		want []*executetest.Block
	}{
		{
			name: `_value>5`,
			spec: &functions.FilterProcedureSpec{
				Fn: &ast.ArrowFunctionExpression{
					Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
					Body: &ast.BinaryExpression{
						Operator: ast.GreaterThanOperator,
						Left: &ast.MemberExpression{
							Object: &ast.Identifier{
								Name: "r",
							},
							Property: &ast.StringLiteral{Value: "_value"},
						},
						Right: &ast.FloatLiteral{
							Value: 5,
						},
					},
				},
			},
			data: []execute.Block{&executetest.Block{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  3,
				},
				ColMeta: []execute.ColMeta{
					{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
					{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0},
					{execute.Time(2), 6.0},
				},
			}},
			want: []*executetest.Block{{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  3,
				},
				ColMeta: []execute.ColMeta{
					{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
					{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
				},
				Data: [][]interface{}{
					{execute.Time(2), 6.0},
				},
			}},
		},
		{
			name: "_value>5 multiple blocks",
			spec: &functions.FilterProcedureSpec{
				Fn: &ast.ArrowFunctionExpression{
					Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
					Body: &ast.BinaryExpression{
						Operator: ast.GreaterThanOperator,
						Left: &ast.MemberExpression{
							Object: &ast.Identifier{
								Name: "r",
							},
							Property: &ast.StringLiteral{Value: "_value"},
						},
						Right: &ast.FloatLiteral{
							Value: 5,
						},
					},
				},
			},
			data: []execute.Block{
				&executetest.Block{
					Bnds: execute.Bounds{
						Start: 1,
						Stop:  3,
					},
					ColMeta: []execute.ColMeta{
						{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
						{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
					},
					Data: [][]interface{}{
						{execute.Time(1), 3.0},
						{execute.Time(2), 6.0},
						{execute.Time(2), 1.0},
					},
				},
				&executetest.Block{
					Bnds: execute.Bounds{
						Start: 3,
						Stop:  5,
					},
					ColMeta: []execute.ColMeta{
						{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
						{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
					},
					Data: [][]interface{}{
						{execute.Time(3), 3.0},
						{execute.Time(3), 2.0},
						{execute.Time(4), 8.0},
					},
				},
			},
			want: []*executetest.Block{
				{
					Bnds: execute.Bounds{
						Start: 1,
						Stop:  3,
					},
					ColMeta: []execute.ColMeta{
						{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
						{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
					},
					Data: [][]interface{}{
						{execute.Time(2), 6.0},
					},
				},
				{
					Bnds: execute.Bounds{
						Start: 3,
						Stop:  5,
					},
					ColMeta: []execute.ColMeta{
						{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
						{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
					},
					Data: [][]interface{}{
						{execute.Time(4), 8.0},
					},
				},
			},
		},
		{
			name: "_value>5 and t1 = a and t2 = y",
			spec: &functions.FilterProcedureSpec{
				Fn: &ast.ArrowFunctionExpression{
					Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
					Body: &ast.LogicalExpression{
						Operator: ast.AndOperator,
						Left: &ast.BinaryExpression{
							Operator: ast.GreaterThanOperator,
							Left: &ast.MemberExpression{
								Object: &ast.Identifier{
									Name: "r",
								},
								Property: &ast.StringLiteral{Value: "_value"},
							},
							Right: &ast.FloatLiteral{
								Value: 5,
							},
						},
						Right: &ast.LogicalExpression{
							Operator: ast.AndOperator,
							Left: &ast.BinaryExpression{
								Operator: ast.EqualOperator,
								Left: &ast.MemberExpression{
									Object: &ast.Identifier{
										Name: "r",
									},
									Property: &ast.StringLiteral{Value: "t1"},
								},
								Right: &ast.StringLiteral{
									Value: "a",
								},
							},
							Right: &ast.BinaryExpression{
								Operator: ast.EqualOperator,
								Left: &ast.MemberExpression{
									Object: &ast.Identifier{
										Name: "r",
									},
									Property: &ast.StringLiteral{Value: "t2"},
								},
								Right: &ast.StringLiteral{
									Value: "y",
								},
							},
						},
					},
				},
			},
			data: []execute.Block{&executetest.Block{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  3,
				},
				ColMeta: []execute.ColMeta{
					{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
					{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
					{Label: "t1", Type: execute.TString, Kind: execute.TagColKind, Common: true},
					{Label: "t2", Type: execute.TString, Kind: execute.TagColKind, Common: false},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0, "a", "x"},
					{execute.Time(2), 6.0, "a", "x"},
					{execute.Time(3), 8.0, "a", "y"},
				},
			}},
			want: []*executetest.Block{{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  3,
				},
				ColMeta: []execute.ColMeta{
					{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
					{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
					{Label: "t1", Type: execute.TString, Kind: execute.TagColKind, Common: true},
					{Label: "t2", Type: execute.TString, Kind: execute.TagColKind, Common: false},
				},
				Data: [][]interface{}{
					{execute.Time(3), 8.0, "a", "y"},
				},
			}},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			executetest.ProcessTestHelper(
				t,
				tc.data,
				tc.want,
				func(d execute.Dataset, c execute.BlockBuilderCache) execute.Transformation {
					f, err := functions.NewFilterTransformation(d, c, tc.spec)
					if err != nil {
						t.Fatal(err)
					}
					return f
				},
			)
		})
	}
}

func TestFilter_PushDown(t *testing.T) {
	spec := &functions.FilterProcedureSpec{
		Fn: &ast.ArrowFunctionExpression{
			Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
			Body: &ast.BinaryExpression{
				Operator: ast.NotEqualOperator,
				Left: &ast.MemberExpression{
					Object: &ast.Identifier{
						Name: "r",
					},
					Property: &ast.StringLiteral{Value: "_measurement"},
				},
				Right: &ast.StringLiteral{Value: "mem"},
			},
		},
	}
	root := &plan.Procedure{
		Spec: new(functions.FromProcedureSpec),
	}
	want := &plan.Procedure{
		Spec: &functions.FromProcedureSpec{
			FilterSet: true,
			Filter: &ast.ArrowFunctionExpression{
				Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
				Body: &ast.BinaryExpression{
					Operator: ast.NotEqualOperator,
					Left: &ast.MemberExpression{
						Object: &ast.Identifier{
							Name: "r",
						},
						Property: &ast.StringLiteral{Value: "_measurement"},
					},
					Right: &ast.StringLiteral{Value: "mem"},
				},
			},
		},
	}

	plantest.PhysicalPlan_PushDown_TestHelper(t, spec, root, false, want)
}

func TestFilter_PushDown_MergeExpressions(t *testing.T) {
	testCases := []struct {
		name string
		spec *functions.FilterProcedureSpec
		root *plan.Procedure
		want *plan.Procedure
	}{
		{
			name: "merge with from",
			spec: &functions.FilterProcedureSpec{
				Fn: &ast.ArrowFunctionExpression{
					Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
					Body: &ast.BinaryExpression{
						Operator: ast.NotEqualOperator,
						Left: &ast.MemberExpression{
							Object: &ast.Identifier{
								Name: "r",
							},
							Property: &ast.StringLiteral{Value: "_measurement"},
						},
						Right: &ast.StringLiteral{Value: "cpu"},
					},
				},
			},
			root: &plan.Procedure{
				Spec: &functions.FromProcedureSpec{
					FilterSet: true,
					Filter: &ast.ArrowFunctionExpression{
						Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
						Body: &ast.BinaryExpression{
							Operator: ast.NotEqualOperator,
							Left: &ast.MemberExpression{
								Object: &ast.Identifier{
									Name: "r",
								},
								Property: &ast.StringLiteral{Value: "_measurement"},
							},
							Right: &ast.StringLiteral{Value: "mem"},
						},
					},
				},
			},
			want: &plan.Procedure{
				Spec: &functions.FromProcedureSpec{
					FilterSet: true,
					Filter: &ast.ArrowFunctionExpression{
						Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
						Body: &ast.LogicalExpression{
							Operator: ast.AndOperator,
							Left: &ast.BinaryExpression{
								Operator: ast.NotEqualOperator,
								Left: &ast.MemberExpression{
									Object: &ast.Identifier{
										Name: "r",
									},
									Property: &ast.StringLiteral{Value: "_measurement"},
								},
								Right: &ast.StringLiteral{Value: "mem"},
							},
							Right: &ast.BinaryExpression{
								Operator: ast.NotEqualOperator,
								Left: &ast.MemberExpression{
									Object: &ast.Identifier{
										Name: "r",
									},
									Property: &ast.StringLiteral{Value: "_measurement"},
								},
								Right: &ast.StringLiteral{Value: "cpu"},
							},
						},
					},
				},
			},
		},
		{
			name: "merge with filter",
			spec: &functions.FilterProcedureSpec{
				Fn: &ast.ArrowFunctionExpression{
					Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
					Body: &ast.BinaryExpression{
						Operator: ast.NotEqualOperator,
						Left: &ast.MemberExpression{
							Object: &ast.Identifier{
								Name: "r",
							},
							Property: &ast.StringLiteral{Value: "_measurement"},
						},
						Right: &ast.StringLiteral{Value: "cpu"},
					},
				},
			},
			root: &plan.Procedure{
				Spec: &functions.FilterProcedureSpec{
					Fn: &ast.ArrowFunctionExpression{
						Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
						Body: &ast.BinaryExpression{
							Operator: ast.NotEqualOperator,
							Left: &ast.MemberExpression{
								Object: &ast.Identifier{
									Name: "r",
								},
								Property: &ast.StringLiteral{Value: "_measurement"},
							},
							Right: &ast.StringLiteral{Value: "mem"},
						},
					},
				},
			},
			want: &plan.Procedure{
				Spec: &functions.FilterProcedureSpec{
					Fn: &ast.ArrowFunctionExpression{
						Params: []*ast.Property{{Key: &ast.Identifier{Name: "r"}}},
						Body: &ast.LogicalExpression{
							Operator: ast.AndOperator,
							Left: &ast.BinaryExpression{
								Operator: ast.NotEqualOperator,
								Left: &ast.MemberExpression{
									Object: &ast.Identifier{
										Name: "r",
									},
									Property: &ast.StringLiteral{Value: "_measurement"},
								},
								Right: &ast.StringLiteral{Value: "mem"},
							},
							Right: &ast.BinaryExpression{
								Operator: ast.NotEqualOperator,
								Left: &ast.MemberExpression{
									Object: &ast.Identifier{
										Name: "r",
									},
									Property: &ast.StringLiteral{Value: "_measurement"},
								},
								Right: &ast.StringLiteral{Value: "cpu"},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			plantest.PhysicalPlan_PushDown_TestHelper(t, tc.spec, tc.root, false, tc.want)
		})
	}
}

//{
//	name: "duplicate non expressions 0",
//	spec: &functions.FilterProcedureSpec{
//		Fn: &ast.ArrowFunctionExpression{
//			Params: []*ast.Property{{Key:&ast.Identifier{Name: "r"}}},
//			Body: &ast.BlockStatement{
//				Body: []ast.Statement{
//					&ast.ReturnStatement{
//						Argument: &ast.BinaryExpression{
//							Operator: ast.NotEqualOperator,
//							Left: &ast.MemberExpression{
//								Object: &ast.Identifier{
//									Name: "r",
//								},
//								Property: &ast.StringLiteral{Value: "_measurement"},
//							},
//							Right: &ast.StringLiteral{Value: "cpu"},
//						},
//					},
//				},
//			},
//		},
//	},
//	root: &plan.Procedure{
//		Spec: &functions.FromProcedureSpec{
//			FilterSet: true,
//			Filter: &ast.ArrowFunctionExpression{
//				Params: []*ast.Property{{Key:&ast.Identifier{Name: "r"}}},
//				Body: &ast.BinaryExpression{
//					Operator: ast.NotEqualOperator,
//					Left: &ast.MemberExpression{
//						Object: &ast.Identifier{
//							Name: "r",
//						},
//						Property: &ast.StringLiteral{Value: "_measurement"},
//					},
//					Right: &ast.StringLiteral{Value: "mem"},
//				},
//			},
//		},
//	},
//	want: &plan.Procedure{
//		Spec: &functions.FromProcedureSpec{
//			FilterSet: false,
//		},
//	},
//	wantDuplicated: true,
//},
//{
//	name: "duplicate non expressions 1",
//	spec: &functions.FilterProcedureSpec{
//		Fn: &ast.ArrowFunctionExpression{
//			Params: []*ast.Property{{Key:&ast.Identifier{Name: "r"}}},
//			Body: &ast.BinaryExpression{
//				Operator: ast.NotEqualOperator,
//				Left: &ast.MemberExpression{
//					Object: &ast.Identifier{
//						Name: "r",
//					},
//					Property: &ast.StringLiteral{Value: "_measurement"},
//				},
//				Right: &ast.StringLiteral{Value: "mem"},
//			},
//		},
//	},
//	root: &plan.Procedure{
//		Spec: &functions.FromProcedureSpec{
//			FilterSet: true,
//			Filter: &ast.ArrowFunctionExpression{
//				Params: []*ast.Property{{Key:&ast.Identifier{Name: "r"}}},
//				Body: &ast.BlockStatement{
//					Body: []ast.Statement{
//						&ast.ReturnStatement{
//							Argument: &ast.BinaryExpression{
//								Operator: ast.NotEqualOperator,
//								Left: &ast.MemberExpression{
//									Object: &ast.Identifier{
//										Name: "r",
//									},
//									Property: &ast.StringLiteral{Value: "_measurement"},
//								},
//								Right: &ast.StringLiteral{Value: "cpu"},
//							},
//						},
//					},
//				},
//			},
//		},
//	},
//	want: &plan.Procedure{
//		Spec: &functions.FromProcedureSpec{
//			FilterSet: false,
//		},
//	},
//	wantDuplicated: true,
//},
