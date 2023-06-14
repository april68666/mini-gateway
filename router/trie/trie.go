package trie

import (
	"strings"
)

type Trie[T comparable] struct {
	node *node[T]
}

type node[T comparable] struct {
	value    T
	path     string
	wildCard bool
	children map[string]*node[T]
}

func NewTrie[T comparable]() *Trie[T] {
	return &Trie[T]{
		node: &node[T]{
			children: make(map[string]*node[T]),
		},
	}
}

func (t *Trie[T]) Insert(path string, value T) {
	n := t.node
	path = strings.Trim(path, "/")
	for _, v := range strings.Split(path, "/") {
		if n.children[v] == nil {
			n.children[v] = &node[T]{
				children: make(map[string]*node[T]),
				path:     v,
			}
		}
		if v == "*" || strings.HasPrefix(v, "{") && strings.HasSuffix(v, "}") {
			n.wildCard = true
		}
		n = n.children[v]
	}
	n.value = value
}

func (t *Trie[T]) Search(path string) (map[string]string, T, bool) {
	var zero T
	n := t.node
	path = strings.Trim(path, "/")
	params := make(map[string]string)
	for _, v := range strings.Split(path, "/") {
		if n.wildCard {
			for k := range n.children {
				if strings.HasPrefix(k, "{") && strings.HasSuffix(k, "}") {
					key := k[1 : len(k)-1]
					params[key] = v
				}
				v = k
			}
		}
		if n.children[v] == nil {
			return nil, zero, false
		}
		n = n.children[v]
	}

	if len(n.children) == 0 {
		return params, n.value, true
	}
	return nil, zero, false
}

func (t *Trie[T]) Delete(path string) {
	path = strings.Trim(path, "/")
	nodes := make([]*node[T], 0)
	n := t.node
	nodes = append(nodes, n)
	for _, v := range strings.Split(path, "/") {
		n = n.children[v]
		nodes = append(nodes, n)
	}
	for i := len(nodes) - 1; i > 0; i-- {
		if len(nodes[i].children) == 0 {
			delete(nodes[i-1].children, nodes[i].path)
		}
	}
}
