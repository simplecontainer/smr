package registry

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"sync"
	"time"
)

const (
	MaxBackoffAttempts  = 5
	RestartWindow       = 10 * time.Minute // Time window to track restarts
	MaxRestartsInWindow = 5                // Max restarts allowed in the window
	MinHealthyRuntime   = 2 * time.Minute  // Container must run this long to be considered "healthy"
)

type RestartTracker struct {
	Count         uint64
	RestartTimes  []time.Time
	LastStartTime time.Time
	IsRunning     bool
}

type Registry struct {
	Containers     map[string]platforms.IContainer
	ContainersLock sync.RWMutex
	BackOffTracker map[string]*RestartTracker
	Client         *clients.Http
	User           *authentication.User
}
