package selector

import "net/http"

func NewNode(scheme, addr, protocol string, weight int, client *http.Client) *Node {
	return &Node{
		scheme:   scheme,
		addr:     addr,
		protocol: protocol,
		weight:   weight,
		client:   client,
	}
}

type Node struct {
	scheme   string
	addr     string
	protocol string
	weight   int
	client   *http.Client
}

func (n *Node) Scheme() string {
	return n.scheme
}

func (n *Node) Address() string {
	return n.addr
}

func (n *Node) Protocol() string {
	return n.protocol
}

func (n *Node) Weight() int {
	return n.weight
}

func (n *Node) Client() *http.Client {
	return n.client
}
