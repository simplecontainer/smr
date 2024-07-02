package watcher

import (
	"context"
	v1 "github.com/qdnqn/smr/pkg/definitions/v1"
	"time"
)

type ContainersWatcher struct {
	Containers map[string]*Containers
}

type Containers struct {
	Definition      v1.Containers
	Syncing         bool
	Tracking        bool
	ContainersQueue chan string        `json:"-"`
	Ctx             context.Context    `json:"-"`
	Cancel          context.CancelFunc `json:"-"`
	Ticker          *time.Ticker       `json:"-"`
}
