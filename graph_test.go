package dag

import "testing"

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

func TestGraph(t *testing.T) {

}
