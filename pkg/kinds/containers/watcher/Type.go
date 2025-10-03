package watcher

import (
	"context"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/dependency"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/readiness"
	"go.uber.org/zap"
	"sync"
	"sync/atomic"
	"time"
)

type Containers struct {
	Watchers map[string]*Container
	Lock     *sync.RWMutex
}

type Container struct {
	Container           platforms.IContainer
	Done                bool
	AllowPlatformEvents bool                           // Will be replaced with atomic version
	ContainerQueue      chan platforms.IContainer      `json:"-"`
	ReadinessChan       chan *readiness.ReadinessState `json:"-"`
	DependencyChan      chan *dependency.State         `json:"-"`
	DeleteC             chan platforms.IContainer      `json:"-"`
	Ctx                 context.Context                `json:"-"`
	Cancel              context.CancelFunc             `json:"-"`
	ReconcileCtx        context.Context                `json:"-"`
	ReconcileCancel     context.CancelFunc             `json:"-"`
	ChecksCtx           context.Context                `json:"-"`
	ChecksCancel        context.CancelFunc             `json:"-"`
	Ticker              *time.Ticker                   `json:"-"`
	Retry               int                            `json:"-"`
	Logger              *zap.Logger
	LogPath             string
	User                *authentication.User `json:"-"`

	mu                  sync.RWMutex `json:"-"`
	allowPlatformEvents atomic.Bool  `json:"-"`
	done                atomic.Bool  `json:"-"`
	tickerMu            sync.Mutex   `json:"-"` // Protects ticker operations
}

func (c *Container) SetAllowPlatformEvents(allow bool) {
	c.allowPlatformEvents.Store(allow)
}

func (c *Container) GetAllowPlatformEvents() bool {
	return c.allowPlatformEvents.Load()
}

func (c *Container) SetDone(done bool) {
	c.done.Store(done)
	if done {
		c.Done = true
	}
}

func (c *Container) IsDone() bool {
	return c.done.Load()
}

func (c *Container) SafeStopTicker() {
	c.tickerMu.Lock()
	defer c.tickerMu.Unlock()

	if c.Ticker != nil {
		c.Ticker.Stop()
	}
}

func (c *Container) SafeResetTicker(d time.Duration) {
	c.tickerMu.Lock()
	defer c.tickerMu.Unlock()

	if c.Ticker != nil {
		c.Ticker.Reset(d)
	}
}

func (c *Container) SendToQueue(container platforms.IContainer, timeout time.Duration) bool {
	if c.IsDone() {
		return false
	}

	select {
	case c.ContainerQueue <- container:
		return true
	case <-time.After(timeout):
		c.Logger.Warn("timeout sending to container queue")
		return false
	case <-c.Ctx.Done():
		return false
	}
}

func (c *Container) SendDelete(container platforms.IContainer, timeout time.Duration) bool {
	if c.IsDone() {
		return false
	}

	select {
	case c.DeleteC <- container:
		return true
	case <-time.After(timeout):
		c.Logger.Warn("timeout sending delete signal")
		return false
	case <-c.Ctx.Done():
		return false
	}
}
