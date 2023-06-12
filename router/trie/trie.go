package trie

import (
	"strings"
)

type Trie[T comparable] struct {
	children map[string]*Trie[T]
	wildCard bool
	value    T
}

func NewTrie[T comparable]() *Trie[T] {
	root := &Trie[T]{}
	root.children = make(map[string]*Trie[T])
	root.wildCard = false
	return root
}

func (t *Trie[T]) Insert(path string, value T) {
	node := t
	path = strings.Trim(path, "/")
	for _, v := range strings.Split(path, "/") {
		if node.children[v] == nil {
			node.children[v] = NewTrie[T]()
		}
		if v == "*" || strings.HasPrefix(v, "{") && strings.HasSuffix(v, "}") {
			node.wildCard = true
		}
		node = node.children[v]
	}
	node.wildCard = true
	node.value = value
}

func (t *Trie[T]) Search(path string) (map[string]string, T, bool) {
	var zero T
	node := t
	path = strings.Trim(path, "/")
	params := make(map[string]string)
	for _, v := range strings.Split(path, "/") {
		if node.wildCard {
			for k := range node.children {
				if strings.HasPrefix(k, "{") && strings.HasSuffix(k, "}") {
					key := k[1 : len(k)-1]
					params[key] = v
				}
				v = k
			}
		}
		if node.children[v] == nil {
			return nil, zero, false
		}
		node = node.children[v]
	}

	if len(node.children) == 0 {
		return params, node.value, node.wildCard
	}
	return nil, zero, false
}
