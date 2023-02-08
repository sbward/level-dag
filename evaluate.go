package dag

import (
	"errors"
	"fmt"
	"log"
	"sync"
)

var ErrMinConcurrency = errors.New("concurrency must be at least 1")

// Evaluate performs a parallel execution of the Graph with the number of workers equal to "concurrency".
// Results can be read directly from each Node after evaluation via the Node.Result field.
func (g Graph) Evaluate(concurrency int) error {
	if concurrency < 1 {
		return ErrMinConcurrency
	}
	nodes, err := g.TopologicalSort()
	if err != nil {
		return fmt.Errorf("topological sort: %w", err)
	}

	log.Printf("evaluation started: concurrency=%d order=%v", concurrency, nodeIDs(nodes))

	// Enqueue nodes in topological order.
	queue := make(chan *Node)
	go func() {
		for _, node := range nodes {
			queue <- node
		}
		close(queue)
	}()

	wait := &sync.WaitGroup{}

	// Launch concurrent workers to evaluate Nodes taken from the queue.
	for i := 0; i < concurrency; i++ {
		wait.Add(1)
		go func(i int) {
			for node := range queue {
				log.Printf("worker %d: evaluating node %s", i, node.ID)
				node.evaluate()
			}
			wait.Done()
		}(i)
	}

	wait.Wait()

	return nil
}

func (n *Node) evaluate() {
	n.wait.Wait()
	close(n.inputs)
	n.Result = n.eval(n.inputs)
	log.Printf("evaluating node %s (%d inputs): result=%d", n.ID, n.indegree, n.Result)
	for _, next := range n.Next {
		next.receive(n.Result)
	}
}

func (n *Node) receive(input int) {
	n.inputs <- input
	n.wait.Done()
}

// Constant returns an EvalFunc that always returns the given integer.
func Constant(n int) EvalFunc {
	return func(_ chan int) int {
		return n
	}
}

// Max is an EvalFunc that returns the highest input or zero if there are no inputs.
func Max(inputs chan int) (output int) {
	for input := range inputs {
		if input > output {
			output = input
		}
	}
	return
}

// Min is an EvalFunc that returns the lowest input or zero if there are no inputs.
func Min(inputs chan int) int {
	output, ok := <-inputs
	if !ok {
		return 0
	}
	for input := range inputs {
		if input < output {
			output = input
		}
	}
	return output
}

// Sum is an EvalFunc that returns the sum of the inputs or zero if there are no inputs.
func Sum(inputs chan int) (output int) {
	for input := range inputs {
		output += input
	}
	return
}
