package rotation

import (
	"context"
	"errors"
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
	return &rotationSelector{index: -1}
}

type rotationSelector struct {
	index int32
	nodes []*selector.Node
}

func (s *rotationSelector) Select(ctx context.Context) (*selector.Node, error) {
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

func (s *rotationSelector) Update(nodes []*selector.Node) {
	s.nodes = nodes
}
