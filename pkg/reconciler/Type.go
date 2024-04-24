package reconciler

import "smr/pkg/container"

type Reconciler struct {
	QueueChan chan Reconcile
}

type Reconcile struct {
	Container *container.Container
}
