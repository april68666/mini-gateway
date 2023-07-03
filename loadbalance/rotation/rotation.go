package rotation

import (
	"context"
	"fmt"
	"mini-gateway/discovery"
	"mini-gateway/loadbalance"
	"mini-gateway/reqcontext"
	"sync/atomic"
)

const NAME = "rotation"

func init() {
	loadbalance.Register(NAME, Factor)
}

func Factor() loadbalance.Picker {
	return newRotationPicker()
}

func newRotationPicker() *rotationPicker {
	return &rotationPicker{}
}

type rotationPicker struct {
	nodes atomic.Value
	// nodes map[string]*node
}

func (s *rotationPicker) Next(ctx context.Context) (discovery.Node, error) {
	color, _ := reqcontext.Color(ctx)
	n := s.nodes.Load().(map[string]*node)[color]
	if n == nil {
		return nil, fmt.Errorf("no node for color:%s was found", color)
	}
	index := n.index.Add(1)
	if index >= int32(len(n.nodes)) {
		n.index.Store(0)
		index = 0
	}
	return n.nodes[index], nil
}

func (s *rotationPicker) Apply(nodes []discovery.Node) {
	ns := make(map[string]*node)
	for _, n := range nodes {
		color, _ := n.Tag("color")
		if v, ok := ns[color]; ok {
			v.nodes = append(v.nodes, n)
		} else {
			newNode := &node{index: atomic.Int32{}}
			newNode.index.Store(-1)
			newNode.nodes = append(newNode.nodes, n)
			ns[color] = newNode
		}
	}
	s.nodes.Store(ns)
}

type node struct {
	index atomic.Int32
	nodes []discovery.Node
}
