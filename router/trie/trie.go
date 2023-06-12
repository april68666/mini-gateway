package trie

import (
	"net/http"
	"strings"
)

type Trie struct {
	children map[string]*Trie
	wildCard bool
	handler  http.Handler
}

func NewTrie() *Trie {
	root := &Trie{}
	root.children = make(map[string]*Trie)
	root.wildCard = false
	return root
}

func (t *Trie) Insert(path string, handler http.Handler) {
	node := t
	path = strings.Trim(path, "/")
	for _, v := range strings.Split(path, "/") {
		if node.children[v] == nil {
			node.children[v] = NewTrie()
		}
		if v == "*" || strings.HasPrefix(v, "{") && strings.HasSuffix(v, "}") {
			node.wildCard = true
		}
		node = node.children[v]
	}
	node.wildCard = true
	node.handler = handler
}

func (t *Trie) Search(path string) (map[string]string, http.Handler, bool) {
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
			return nil, nil, false
		}
		node = node.children[v]
	}

	if len(node.children) == 0 {
		return params, node.handler, node.wildCard
	}
	return nil, nil, false
}
