package watcher

import (
	"context"
	"github.com/simplecontainer/smr/implementations/container/container"
	"time"
)

type ContainerWatcher struct {
	Container map[string]*Container
}

type Container struct {
	Container      *container.Container
	Syncing        bool
	Tracking       bool
	ContainerQueue chan *container.Container `json:"-"`
	Ctx            context.Context           `json:"-" `
	Cancel         context.CancelFunc        `json:"-"`
	Ticker         *time.Ticker              `json:"-"`
}
