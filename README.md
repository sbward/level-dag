# level-dag

![Coverage](https://img.shields.io/badge/Coverage-97.4%25-brightgreen)
[![Go Reference](https://pkg.go.dev/badge/github.com/sbward/level-dag.svg)](https://pkg.go.dev/github.com/sbward/level-dag)

Package dag implements types and functions to define, construct, and evaluate a connected directed acyclic graph (DAG), where each node is a computation.

## Usage

To construct a Graph, create Nodes with the NewNode function, then pass the head Nodes to the New function.

```go
package main
import "github.com/sbward/level-dag"

func ExampleGraph() (dag.Graph, error) {
	sum := dag.NewNode("sum", dag.Sum)
	max := dag.NewNode("max", dag.Max, sum)
	min := dag.NewNode("min", dag.Min, sum)
	return dag.New(
		dag.NewNode("1", dag.Constant(1), max),
		dag.NewNode("2", dag.Constant(2), max),
		dag.NewNode("3", dag.Constant(3), min),
		dag.NewNode("4", dag.Constant(4), min),
	)
}
```

This function will produce the following Graph:

![Example Graph](assignment_graph.png?raw=true "Example Graph")

### Evaluation

To evaluate a Graph, use the Graph.Evaluate() function, while passing in the desired concurrency.

```go

graph, err := ExampleGraph()
if err != nil {
	return err
}

err := graph.Evaluate(4)
if err != nil {
	return err
}

// Inspect the Graph to view computation results

fmt.Println(graph["max"].Result) // 2
fmt.Println(graph["min"].Result) // 3
fmt.Println(graph["sum"].Result) // 5
```
