package selector

type Selector interface {
	Select() (*Node, error)
	Update(nodes []*Node)
}
