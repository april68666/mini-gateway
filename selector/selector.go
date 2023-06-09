package selector

import (
	"sync"
)

var selectorFactory = sync.Map{}

type Selector interface {
	Select() (*Node, error)
	Update(nodes []*Node)
}

type Factory func() Selector

func Register(name string, f Factory) {
	selectorFactory.Store(name, f)
}

func Get(name string) (f Factory, ok bool) {
	v, ok := selectorFactory.Load(name)
	if !ok {
		return nil, ok
	}
	return v.(Factory), ok
}
