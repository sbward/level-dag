# level-dag

![Coverage](https://img.shields.io/badge/Coverage-97.4%25-brightgreen)
[![Go Reference](https://pkg.go.dev/badge/github.com/sbward/level-dag.svg)](https://pkg.go.dev/github.com/sbward/level-dag)

Package dag implements types and functions to define, construct, and evaluate a connected directed acyclic graph (DAG), where each node is a computation.

## Usage

To construct a `Graph`, create `Node` values with the `NewNode` function, then pass the head `Node` values to the `New` function.

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

The above function will produce the following `Graph`:

![Example Graph](assignment_graph.png?raw=true "Example Graph")

### Evaluation

To evaluate a `Graph`, use the `Graph.Evaluate` function, while passing in the desired concurrency.

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

## Implementation

Each `Node` of a `Graph` has an `Inputs` channel and a `*sync.WaitGroup` for concurrency control. When constructing `Node` values, downstream `Node` values increase their `*sync.WaitGroup` counter by 1 for each `Node` that will provide an input.

When evaluation is started by invoking `Graph.Evaluate`, the head `Node` values are evaluated first. A worker pool receives `Node` references to process in topological order, ensuring that the `Graph` can be processed to completion by any number of concurrent workers.

During evaluation, each `Node` waits for its parents to complete first by waiting for the `*sync.WaitGroup`. Head nodes have no inputs, so this stage is effectively skipped. The `Input` channel is closed, and then the buffered inputs are processed by the `EvalFunc` of the `Node` to merge the inputs into a single result. Finally the `Node` invokes the `Receive` method on each downsteam Node to send its output to the next Node's `Input` channel, and decreases that Node's `*sync.WaitGroup` counter by 1. This process repeats concurrently across all `Nodes` in topological order until the entire `Graph` has been processed.
