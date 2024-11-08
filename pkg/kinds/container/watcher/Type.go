package watcher

import (
	"context"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/dependency"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/readiness"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/types"
	"go.uber.org/zap"
	"time"
)

type ContainerWatcher struct {
	Container    map[string]*Container
	EventChannel chan *types.Events
}

type Container struct {
	Container      platforms.IContainer
	Syncing        bool
	ContainerQueue chan platforms.IContainer      `json:"-"`
	ReadinessChan  chan *readiness.ReadinessState `json:"-"`
	DependencyChan chan *dependency.State         `json:"-"`
	Ctx            context.Context                `json:"-" `
	Cancel         context.CancelFunc             `json:"-"`
	Ticker         *time.Ticker                   `json:"-"`
	Logger         *zap.Logger
	User           *authentication.User `json:"-"`
}
