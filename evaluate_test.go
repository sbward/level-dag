package dag

import (
	"errors"
	"fmt"
	"testing"
)

func assignmentGraph() (Graph, error) {
	sum := NewNode("sum", Sum)
	max := NewNode("max", Max, sum)
	min := NewNode("min", Min, sum)
	return New(
		NewNode("1", Constant(1), max),
		NewNode("2", Constant(2), max),
		NewNode("3", Constant(3), min),
		NewNode("4", Constant(4), min),
	)
}

var evaluateCases = []struct {
	Name           string
	Graph          func() (Graph, error)
	MaxConcurrency int
	ExpectError    error
	ExpectResults  map[string]int // Node ID -> Result
}{
	{
		Name: "empty",
		Graph: func() (Graph, error) {
			return New()
		},
		MaxConcurrency: 1,
	},
	{
		Name:           "assignment",
		Graph:          assignmentGraph,
		MaxConcurrency: 6,
		ExpectResults: map[string]int{
			"1":   1,
			"2":   2,
			"3":   3,
			"4":   4,
			"min": 3,
			"max": 2,
			"sum": 5,
		},
	},
	{
		Name: "split ending",
		Graph: func() (Graph, error) {
			sum := NewNode("sum", Sum)
			min := NewNode("min", Min)
			max := NewNode("max", Max)
			one := NewNode("1", Constant(1), sum, min, max)
			two := NewNode("2", Constant(2), sum, min, max)
			return New(one, two)
		},
		MaxConcurrency: 3,
		ExpectResults: map[string]int{
			"sum": 3,
			"min": 1,
			"max": 2,
		},
	},
	{
		Name: "linked constants",
		Graph: func() (Graph, error) {
			one := NewNode("1", Constant(1))
			two := NewNode("2", Constant(2), one)
			return New(two)
		},
		MaxConcurrency: 3,
		ExpectResults: map[string]int{
			"1": 1,
			"2": 2,
		},
	},
	{
		Name: "no input min",
		Graph: func() (Graph, error) {
			return New(NewNode("min", Min))
		},
		MaxConcurrency: 3,
		ExpectResults: map[string]int{
			"min": 0,
		},
	},
	{
		Name: "no input max",
		Graph: func() (Graph, error) {
			return New(NewNode("max", Max))
		},
		MaxConcurrency: 3,
		ExpectResults: map[string]int{
			"max": 0,
		},
	},
	{
		Name: "no input sum",
		Graph: func() (Graph, error) {
			return New(NewNode("sum", Sum))
		},
		MaxConcurrency: 3,
		ExpectResults: map[string]int{
			"sum": 0,
		},
	},
}

// TestEvaluate runs each test case at every concurrency level from 1 to test.MaxConcurrency.
func TestEvaluate(t *testing.T) {
	for i, test := range evaluateCases {
		for c := 1; c < test.MaxConcurrency; c++ {
			t.Run(fmt.Sprintf("%d_%s_c=%d", i, test.Name, c), func(t *testing.T) {
				graph, err := test.Graph()
				if err != nil && !errors.Is(err, test.ExpectError) {
					t.Fatalf("unexpected error from calling Graph(): %s", err)
				}
				err = graph.Evaluate(c)
				if err != nil && !errors.Is(err, test.ExpectError) {
					t.Fatalf("unexpected error from calling Evaluate(): %s", err)
				}
				// Assert expected results match the results for each Node.
				for id, expected := range test.ExpectResults {
					if result := graph[id].Result; result != expected {
						t.Fatalf("unexpected result for node %s: want %d but got %d", id, expected, result)
					}
				}
			})
		}
	}
}

func TestMinConcurrency(t *testing.T) {
	graph, err := assignmentGraph()
	if err != nil {
		t.Fatal(err)
	}
	if err := graph.Evaluate(0); !errors.Is(err, ErrMinConcurrency) {
		t.Fail()
	}
}
