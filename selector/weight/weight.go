package weight

import (
	"context"
	"math/rand"
	"mini-gateway/reqcontext"
	"mini-gateway/selector"
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
	nodes map[string]*node
}

func (s *weight) Select(ctx context.Context) (*selector.Node, error) {
	color, _ := reqcontext.Color(ctx)
	n := s.nodes[color]
	index := rand.Intn(len(n.nodes))
	return n.nodes[index], nil
}

func (s *weight) Update(nodes []*selector.Node) {
	ns := make(map[string]*node)
	for _, n := range nodes {
		for i := 0; i < n.Weight(); i++ {
			if v, ok := ns[n.Color()]; ok {
				v.nodes = append(v.nodes, n)
			} else {
				newNode := &node{}
				newNode.nodes = append(newNode.nodes, n)
				ns[n.Color()] = newNode
			}
		}
	}
	s.nodes = ns
}

type node struct {
	nodes []*selector.Node
}
