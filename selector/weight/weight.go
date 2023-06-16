package weight

import (
	"context"
	"math/rand"
	"mini-gateway/discovery"
	"mini-gateway/reqcontext"
	"mini-gateway/selector"
	"sync/atomic"
)

const NAME = "weight"

func init() {
	selector.Register(NAME, Factor)
}

func Factor() selector.Selector {
	return newWeightSelector()
}

func newWeightSelector() *weight {
	return &weight{}
}

type weight struct {
	nodes atomic.Value
	// nodes map[string]*node
}

func (s *weight) Select(ctx context.Context) (discovery.Node, error) {
	color, _ := reqcontext.Color(ctx)
	n := s.nodes.Load().(map[string]*node)[color]
	index := rand.Intn(len(n.nodes))
	return n.nodes[index], nil
}

func (s *weight) Apply(nodes []discovery.Node) {
	ns := make(map[string]*node)
	for _, n := range nodes {
		for i := 0; i < n.Weight(); i++ {
			color, _ := n.Tag("color")
			if v, ok := ns[color]; ok {
				v.nodes = append(v.nodes, n)
			} else {
				newNode := &node{}
				newNode.nodes = append(newNode.nodes, n)
				ns[color] = newNode
			}
		}
	}
	s.nodes.Store(ns)
}

type node struct {
	nodes []discovery.Node
}
