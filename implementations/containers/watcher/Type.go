package watcher

import (
	"context"
	"github.com/simplecontainer/smr/pkg/authentication"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"go.uber.org/zap"
	"time"
)

type ContainersWatcher struct {
	Containers map[string]*Containers
}

type Containers struct {
	Definition      v1.ContainersDefinition
	Syncing         bool
	Tracking        bool
	ContainersQueue chan string          `json:"-"`
	User            *authentication.User `json:"-"`
	Ctx             context.Context      `json:"-"`
	Cancel          context.CancelFunc   `json:"-"`
	Ticker          *time.Ticker         `json:"-"`
	Logger          *zap.Logger
}
