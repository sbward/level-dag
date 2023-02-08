package dag

import (
	"fmt"
	"testing"
)

var topologicalSortCases = []struct {
	Name         string
	Graph        func() (Graph, error)
	ExpectError  error
	ExpectResult []string
}{
	{
		Name: "empty",
		Graph: func() (Graph, error) {
			return New()
		},
		ExpectResult: []string{},
	},
	{
		Name: "one node",
		Graph: func() (Graph, error) {
			return New(
				NewNode("1", Constant(1)),
			)
		},
		ExpectResult: []string{"1"},
	},
	{
		Name: "two nodes",
		Graph: func() (Graph, error) {
			return New(
				NewNode("1", Constant(1),
					NewNode("max", Max),
				),
			)
		},
		ExpectResult: []string{"1", "max"},
	},
	{
		Name:  "assignment",
		Graph: assignmentGraph,
	},
}

func TestTopologicalSort(t *testing.T) {
	for i, test := range topologicalSortCases {
		t.Run(fmt.Sprintf("%d_%s", i, test.Name), func(t *testing.T) {
			graph, err := test.Graph()
			if err != nil && err != test.ExpectError {
				t.Fatalf("unexpected error from calling Graph(): %s", err)
			}
			sorted, err := graph.TopologicalSort()
			if err != nil && err != test.ExpectError {
				t.Fatalf("unexpected error from calling TopologicalSort(): %s", err)
			}

			ids := make([]string, len(sorted))
			for i, node := range sorted {
				ids[i] = node.ID
			}
			t.Log("Result:", ids)

			// If a specific result is expected, assert that it matches.
			if test.ExpectResult != nil {
				if len(sorted) != len(test.ExpectResult) {
					t.Fatalf("expected result to have length %d but got %d", len(test.ExpectResult), len(sorted))
				}
				for i, expectNodeID := range test.ExpectResult {
					if nodeID := sorted[i].ID; nodeID != expectNodeID {
						t.Fatalf("unexpected sorting result: want %s at position %d but got %s", expectNodeID, i, nodeID)
					}
				}
			}

			// Assert that the result is a topological ordering.
			// For each node, we check that every dependency was visited first.

			// Create a list of the dependencies for each node.
			deps := map[string]map[string]struct{}{}
			for id := range graph {
				deps[id] = make(map[string]struct{})
			}
			graph.Walk(func(current *Node, prev []*Node) error {
				for _, dep := range prev {
					deps[current.ID][dep.ID] = struct{}{}
				}
				return nil
			})

			// Record whether each Node has been visited.
			visited := map[string]bool{}
			for id := range graph {
				visited[id] = false
			}

			// Step through the topological ordering while marking each node as visited,
			// and checking if the deps of each node have also been visited yet.
			for _, node := range sorted {
				for dep := range deps[node.ID] {
					if _, ok := visited[dep]; !ok {
						t.Fatalf("invalid result order: node %s should come before node %s", dep, node.ID)
					}
				}
				visited[node.ID] = true
			}
		})
	}
}
