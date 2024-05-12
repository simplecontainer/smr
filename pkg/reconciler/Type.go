package reconciler

import "github.com/qdnqn/smr/pkg/container"

type Reconciler struct {
	QueueChan   chan Reconcile
	QueueEvents chan Events
}

type Reconcile struct {
	Container *container.Container
}

type Events struct {
	Kind      string
	Message   string
	Container *container.Container
}
