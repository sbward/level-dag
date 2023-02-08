package dag

import (
	"errors"
	"fmt"
	"log"
	"sync"
)

var ErrMinConcurrency = errors.New("concurrency must be at least 1")

func (g Graph) Evaluate(concurrency int) error {
	if concurrency < 1 {
		return ErrMinConcurrency
	}
	nodes, err := g.TopologicalSort()
	if err != nil {
		return fmt.Errorf("topological sort: %w", err)
	}

	log.Println("evaluation order:", nodeIDs(nodes))

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
		go func() {
			for node := range queue {
				node.Evaluate()
			}
			wait.Done()
		}()
	}

	wait.Wait()

	return nil
}

func (n *Node) Evaluate() {
	n.wait.Wait()
	close(n.inputs)
	n.Result = n.eval(n.inputs)
	log.Printf("evaluating node %s (%d inputs): result=%d", n.ID, n.indegree, n.Result)
	for _, next := range n.Next {
		next.Receive(n.Result)
	}
}

func (n *Node) Receive(input int) {
	n.inputs <- input
	n.wait.Done()
}

func Constant(n int) EvalFunc {
	return func(_ chan int) int {
		return n
	}
}

func Max(inputs chan int) (output int) {
	for input := range inputs {
		if input > output {
			output = input
		}
	}
	return
}

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

func Sum(inputs chan int) (output int) {
	for input := range inputs {
		output += input
	}
	return
}
