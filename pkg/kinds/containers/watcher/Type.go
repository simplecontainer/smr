package watcher

import (
	"context"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/dependency"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/readiness"
	"go.uber.org/zap"
	"sync"
	"time"
)

type Containers struct {
	Watchers map[string]*Container
	Lock     *sync.RWMutex
}

type Container struct {
	Container           platforms.IContainer
	Done                bool
	AllowPlatformEvents bool
	ContainerQueue      chan platforms.IContainer      `json:"-"`
	ReadinessChan       chan *readiness.ReadinessState `json:"-"`
	DependencyChan      chan *dependency.State         `json:"-"`
	PauseC              chan platforms.IContainer      `json:"-"`
	Ctx                 context.Context                `json:"-" `
	Cancel              context.CancelFunc             `json:"-"`
	Ticker              *time.Ticker                   `json:"-"`
	Retry               int                            `json:"-"`
	Logger              *zap.Logger
	LogPath             string
	User                *authentication.User `json:"-"`
}
