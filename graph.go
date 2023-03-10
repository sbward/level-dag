package dag

import (
	"errors"
	"log"
	"sync"
)

// Node is a single computation step in a Graph.
// To construct Nodes, use the NewNode function.
type Node struct {
	ID       string
	Next     []*Node
	Result   int
	eval     EvalFunc
	wait     *sync.WaitGroup
	indegree int
	inputs   chan int
}

// NewNode returns a Node with the given ID and EvalFunc.
// The Node's output will be sent to any Nodes provided as the "next" argument.
func NewNode(id string, eval EvalFunc, next ...*Node) *Node {
	for _, next := range next {
		next.wait.Add(1)
		next.indegree++
	}
	return &Node{
		ID:     id,
		Next:   next,
		eval:   eval,
		wait:   &sync.WaitGroup{},
		inputs: make(chan int, MaxIndegree),
	}
}

// MaxIndegree sets the buffer size of the Inputs channel for Nodes.
var MaxIndegree = 10

// EvalFunc accepts a channel of zero or more numerical inputs and returns a single numerical output.
type EvalFunc func(chan int) int

// Graph is a directed acyclic graph of Nodes. Map keys are Node IDs.
type Graph map[string]*Node

// New constructs a Graph from the given Nodes.
// Only head Nodes need to be passed to New; these Nodes will be traversed and connected to form the full Graph.
// Each Node must have a unique ID.
// If the Graph contains a cycle, ErrCycle is returned.
// If one or more Nodes have no path to the rest of the Nodes, ErrDisconnected is returned.
func New(nodes ...*Node) (Graph, error) {
	g := Graph(make(map[string]*Node, len(nodes)))

	// Add every Node to the Graph while checking for cycles.
	for _, node := range nodes {
		err := node.walkRecursive(func(current *Node, prev []*Node) error {
			for _, p := range prev {
				// If the Node was already visited in prev, there is a cycle.
				if current.ID == p.ID {
					log.Printf("cycle: node %s is referenced by descendent node %s", p.ID, current.ID)
					return ErrCycle
				}
			}
			if _, ok := g[current.ID]; ok {
				// Node was already recorded, ok to skip.
				return nil
			}
			g[current.ID] = current
			return nil
		}, []*Node{})

		if err != nil {
			return nil, err
		}
	}

	// Check connectivity.
	if err := g.CheckConnectivity(); err != nil {
		return nil, err
	}

	return g, nil
}

// ErrCycle is returned when a cycle is detected in a Graph.
var ErrCycle = errors.New("cycle detected")

// ErrDisconnected is returned when a Node is unreachable from at least one Node in the same Graph.
var ErrDisconnected = errors.New("disconnected node")

// CheckConnectivity returns ErrDisconnect if the Graph is disconnected.
func (g Graph) CheckConnectivity() error {
	var connected = map[string]map[string]bool{}

	// Initialize a connectivity map that records whether a Node connects to each other Node.
	// The structure of the map is [Node A ID] -> [Node B ID] -> Is Connected (bool).
	for _, src := range g {
		connected[src.ID] = make(map[string]bool)
	inner:
		for _, dst := range g {
			if dst.ID == src.ID {
				log.Printf("skipping %s to %s", src.ID, dst.ID)
				continue inner
			}
			log.Printf("init connection: %s to %s", src.ID, dst.ID)
			connected[src.ID][dst.ID] = false
		}
	}

	// Traverse the Graph depth-first to check for cycles while recording connectivity.
	g.Walk(func(current *Node, prev []*Node) error {
		for _, p := range prev {
			// Mark each previously visited Node as connected to this Node and its connections, and vice versa.
			log.Printf("connected: %s to %s", current.ID, p.ID)
			connected[current.ID][p.ID] = true
			connected[p.ID][current.ID] = true
			for connID, ok := range connected[current.ID] {
				if ok {
					connected[p.ID][connID] = true
				}
			}
			for connID, ok := range connected[p.ID] {
				if ok {
					connected[current.ID][connID] = true
				}
			}
		}
		return nil
	})

	reversed := g.Reversed()

	// For every Node in the reversed graph, complete the connectivity check by doing
	// another depth-first traversal and marking all Nodes reached.
	reversed.Walk(func(current *Node, prev []*Node) error {
		for _, p := range prev {
			connected[current.ID][p.ID] = true
			connected[p.ID][current.ID] = true
			for connID, ok := range connected[current.ID] {
				if ok {
					connected[p.ID][connID] = true
				}
			}
			for connID, ok := range connected[p.ID] {
				if ok {
					connected[current.ID][connID] = true
				}
			}
		}
		return nil
	})

	// If any Nodes have not reached any other Nodes, return ErrDisconnected.
	for src, dst := range connected {
		for dst, reached := range dst {
			if !reached {
				log.Printf("disconnect: node %s is not connected to node %s", src, dst)
				return ErrDisconnected
			}
		}
	}

	return nil
}

// Filter returns the Nodes in the graph that pass the given filter check.
func (g Graph) Filter(filter func(*Node) bool) []*Node {
	out := make([]*Node, 0)
	for _, n := range g {
		if filter(n) {
			out = append(out, n)
		}
	}
	return out
}

// Roots returns the root Nodes of the Graph (Nodes with indegree of 0).
func (g Graph) Roots() []*Node {
	return g.Filter(func(n *Node) bool { return n.indegree == 0 })
}

// Walk recursively traverses the Graph depth-first, applying the visit function to each visited Node.
// The visit function also receives the chain of Nodes visited prior to the current Node,
// sorted so that the root is at index 0 of the slice, and the previously visited Node is at the end of the slice.
func (g Graph) Walk(visit func(current *Node, prev []*Node) error) error {
	for _, n := range g.Roots() {
		if err := n.walkRecursive(visit, []*Node{}); err != nil {
			return err
		}
	}
	return nil
}

func (n *Node) walkRecursive(visit func(current *Node, prev []*Node) error, prev []*Node) error {
	if err := visit(n, prev); err != nil {
		return err
	}
	for _, next := range n.Next {
		if err := next.walkRecursive(visit, append(prev, n)); err != nil {
			return err
		}
	}
	return nil
}

// Reversed returns a new Graph with the edge directions reversed.
func (g Graph) Reversed() Graph {
	result := make(Graph)
	g.Walk(func(current *Node, prev []*Node) error {
		// Add a copy of the Node to the reversed Graph without any edges if we haven't done so yet.
		if _, ok := result[current.ID]; !ok {
			result[current.ID] = &Node{
				ID:     current.ID,
				Next:   []*Node{},
				eval:   current.eval,
				wait:   &sync.WaitGroup{},
				inputs: make(chan int),
			}
		}
		// If the current Node has no parent, continue.
		if len(prev) == 0 {
			return nil
		}
		// Connect the copy of the current Node to the copy of the parent Node if we haven't done so yet.
		parent := prev[len(prev)-1]
		for _, next := range result[current.ID].Next {
			if next.ID == parent.ID {
				// Already connected; continue walking.
				return nil
			}
		}
		result[current.ID].Next = append(result[current.ID].Next, result[parent.ID])
		return nil
	})
	return result
}
