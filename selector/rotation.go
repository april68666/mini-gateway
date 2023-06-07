package selector

import (
	"errors"
	"sync/atomic"
)

func NewRotationSelector() *RotationSelector {
	return &RotationSelector{index: -1}
}

type RotationSelector struct {
	index int32
	nodes []*Node
}

func (s *RotationSelector) Select() (*Node, error) {
	nodes := s.nodes
	if len(nodes) == 0 {
		return nil, errors.New("node not found")
	}
	index := atomic.AddInt32(&s.index, 1)
	if index >= int32(len(nodes)) {
		atomic.StoreInt32(&s.index, 0)
		index = 0
	}
	return nodes[index], nil
}

func (s *RotationSelector) Update(nodes []*Node) {
	s.nodes = nodes
}
