package events

import "github.com/simplecontainer/container/container"

type Reconcile struct {
	Container *container.Container
}

type Events struct {
	Kind      string
	Message   string
	Container *container.Container
}
