package dag

import (
	"errors"
	"fmt"
	"testing"
)

var graphTestCases = []struct {
	Name        string
	Graph       func() (Graph, error)
	ExpectError error
}{
	{
		Name: "cycle",
		Graph: func() (Graph, error) {
			a, b := NewNode("a", Constant(1)), NewNode("b", Constant(2))
			a.Next = append(a.Next, b)
			b.Next = append(b.Next, a)
			return New(a, b)
		},
		ExpectError: ErrCycle,
	},
	{
		Name: "disconnect",
		Graph: func() (Graph, error) {
			a, b := NewNode("a", Constant(1)), NewNode("b", Constant(2))
			return New(a, b)
		},
		ExpectError: ErrDisconnected,
	},
}

func TestGraph(t *testing.T) {
	for i, test := range graphTestCases {
		t.Run(fmt.Sprintf("%d_%s", i, test.Name), func(t *testing.T) {
			_, err := test.Graph()
			if err != nil && !errors.Is(err, test.ExpectError) {
				t.Fatalf("unexpected error from calling Graph(): %s", err)
			}
		})
	}
}
