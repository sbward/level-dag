package dag

import (
	"fmt"
	"testing"
)

var evaluateCases = []struct {
	Name          string
	Graph         func() (Graph, error)
	Concurrency   int
	ExpectError   error
	ExpectResults map[string]int // Node ID -> Result
}{
	{
		Name: "empty",
		Graph: func() (Graph, error) {
			return New()
		},
		Concurrency: 1,
	},
	{
		Name:        "assignment",
		Graph:       assignmentGraph,
		Concurrency: 1,
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
}

func TestEvaluate(t *testing.T) {
	for i, test := range evaluateCases {
		t.Run(fmt.Sprintf("%d_%s", i, test.Name), func(t *testing.T) {
			graph, err := test.Graph()
			if err != nil && err != test.ExpectError {
				t.Fatalf("unexpected error from calling Graph(): %s", err)
			}
			err = graph.Evaluate(test.Concurrency)
			if err != nil && err != test.ExpectError {
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
