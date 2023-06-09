package weight

import (
	"context"
	"errors"
	"math/rand"
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
	nodes []*selector.Node
}

func (s *weight) Select(ctx context.Context) (*selector.Node, error) {
	nodes := s.nodes
	if len(nodes) == 0 {
		return nil, errors.New("node not found")
	}
	index := rand.Intn(len(nodes))
	return nodes[index], nil
}

func (s *weight) Update(nodes []*selector.Node) {
	ns := make([]*selector.Node, 0)
	for _, node := range nodes {
		for i := 0; i < node.Weight(); i++ {
			ns = append(ns, node)
		}
	}
	s.nodes = ns
}
