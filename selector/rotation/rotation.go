package rotation

import (
	"context"
	"mini-gateway/reqcontext"
	"mini-gateway/selector"
	"sync/atomic"
)

const NAME = "rotation"

func init() {
	selector.Register(NAME, Factor)
}

func Factor() selector.Selector {
	return newRotationSelector()
}

func newRotationSelector() *rotationSelector {
	return &rotationSelector{}
}

type rotationSelector struct {
	nodes atomic.Value
	// nodes map[string]*node
}

func (s *rotationSelector) Select(ctx context.Context) (*selector.Node, error) {
	color, _ := reqcontext.Color(ctx)
	n := s.nodes.Load().(map[string]*node)[color]
	index := atomic.AddInt32(&n.index, 1)
	if index >= int32(len(n.nodes)) {
		atomic.StoreInt32(&n.index, 0)
		index = 0
	}
	return n.nodes[index], nil
}

func (s *rotationSelector) Update(nodes []*selector.Node) {
	ns := make(map[string]*node)
	for _, n := range nodes {
		if v, ok := ns[n.Color()]; ok {
			v.nodes = append(v.nodes, n)
		} else {
			newNode := &node{index: -1}
			newNode.nodes = append(newNode.nodes, n)
			ns[n.Color()] = newNode
		}
	}
	s.nodes.Store(ns)
}

type node struct {
	index int32
	nodes []*selector.Node
}
