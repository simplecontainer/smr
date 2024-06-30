package reconciler

import (
	"context"
	"github.com/qdnqn/smr/pkg/container"
	v1 "github.com/qdnqn/smr/pkg/definitions/v1"
	"time"
)

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

type ContainersWatcher struct {
	Containers map[string]*Containers
}

type Containers struct {
	Definition     *v1.Containers
	InSync         bool
	ContainerQueue chan Event      `json:"-"`
	Ctx            context.Context `json:"-"`
	Ticker         *time.Ticker    `json:"-"`
}

type Event struct {
	Event string
}
