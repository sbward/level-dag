package dag

// TopologicalSort returns a slice containing every Node in the Graph sorted in an order
// which guarantees that each node is placed after any Nodes that it depends upon in the Graph.
// If a cycle is detected during iteration, ErrCycle is returned.
func (g Graph) TopologicalSort() ([]*Node, error) {
	s := &topologicalSort{
		visiting: make(map[*Node]struct{}),
		visited:  make(map[*Node]struct{}),
		sorted:   make([]*Node, 0),
	}

	// Begin topological sorting by visiting each Node with indegree 0 (roots).
	for _, node := range g.Roots() {
		if err := s.visit(node); err != nil {
			return nil, err
		}
	}

	// Return a slice containing Nodes in topological order.
	return s.sorted, nil
}

type topologicalSort struct {
	visiting, visited map[*Node]struct{}
	sorted            []*Node
}

func (s *topologicalSort) prependToSorted(n *Node) {
	s.sorted = append([]*Node{n}, s.sorted...)
}

func (s *topologicalSort) visit(node *Node) error {
	// If the node is visited, return.
	if _, ok := s.visited[node]; ok {
		return nil
	}

	// If the node is visiting, there is a cycle in the graph.
	if _, ok := s.visiting[node]; ok {
		return ErrCycle
	}

	// Mark the node as visiting ("temporary mark").
	s.visiting[node] = struct{}{}

	// Visit each "next" node (nodes that depend on this one).
	for _, next := range node.Next {
		s.visit(next)
	}

	// Unmark the node as visiting.
	delete(s.visiting, node)

	// Mark the node as visited ("permanent mark").
	s.visited[node] = struct{}{}

	// Prepend node to the sorted list.
	s.prependToSorted(node)

	return nil
}

func nodeIDs(nodes []*Node) []string {
	out := make([]string, len(nodes))
	for i, node := range nodes {
		out[i] = node.ID
	}
	return out
}
