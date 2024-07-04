package events

import "github.com/simplecontainer/smr/implementations/container/container"

type Reconcile struct {
	Container *container.Container
}

type Events struct {
	Kind      string
	Message   string
	Container *container.Container
}
