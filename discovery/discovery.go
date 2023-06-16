package discovery

import (
	"context"
)

type Resolver interface {
	Resolve(ctx context.Context, desc string) (*Result, error)
	Watch(ctx context.Context, desc string, callBack func(*Result)) error
}

type Result struct {
	Nodes []Node
}

type Node interface {
	Uri() string
	Weight() int
	Tag(key string) (value string, exist bool)
}

func NewNode(uri string, weight int, tags map[string]string) Node {
	return &node{
		uri:    uri,
		weight: weight,
		tags:   tags,
	}
}

type node struct {
	uri    string
	weight int
	tags   map[string]string
}

func (n *node) Uri() string {
	return n.uri
}

func (n *node) Weight() int {
	return n.weight
}

func (n *node) Tag(key string) (value string, exist bool) {
	value, exist = n.tags[key]
	return
}
