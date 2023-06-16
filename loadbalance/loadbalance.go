package loadbalance

import (
	"context"
	"mini-gateway/discovery"
)

type LoadBalancer interface {
	GetPicker() Picker
}

type Picker interface {
	Next(ctx context.Context) discovery.Node
}
