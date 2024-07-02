package reconciler

import "github.com/qdnqn/container"

type Reconciler struct {
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
