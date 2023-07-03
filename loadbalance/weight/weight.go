package weight

import (
	"context"
	"fmt"
	"math/rand"
	"mini-gateway/discovery"
	"mini-gateway/loadbalance"
	"mini-gateway/reqcontext"
	"sync/atomic"
)

const NAME = "weight"

func init() {
	loadbalance.Register(NAME, Factor)
}

func Factor() loadbalance.Picker {
	return newWeightPicker()
}

func newWeightPicker() *weight {
	return &weight{}
}

type weight struct {
	nodes atomic.Value
	// nodes map[string]*node
}

func (s *weight) Next(ctx context.Context) (discovery.Node, error) {
	color, _ := reqcontext.Color(ctx)
	n := s.nodes.Load().(map[string]*node)[color]
	if n == nil {
		return nil, fmt.Errorf("no node for color:%s was found", color)
	}
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
