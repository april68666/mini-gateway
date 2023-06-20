package loadbalance

import (
	"context"
	"mini-gateway/discovery"
	"sync"
)

var pickerFactory = sync.Map{}

type Picker interface {
	Next(ctx context.Context) (discovery.Node, error)
	Apply(nodes []discovery.Node)
}

type Factory func() Picker

func Register(name string, f Factory) {
	pickerFactory.Store(name, f)
}

func GetPicker(name string) (f Factory, ok bool) {
	v, ok := pickerFactory.Load(name)
	if !ok {
		return nil, ok
	}
	return v.(Factory), ok
}
