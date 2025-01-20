package watcher

import (
	"context"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/implementation"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/registry"
	"go.uber.org/zap"
	"time"
)

type RepositoryWatcher struct {
	Repositories map[string]*Gitops
}

type Gitops struct {
	Gitops      *implementation.Gitops
	Registry    *registry.Registry
	Children    []*common.Request
	User        *authentication.User        `json:"-"`
	GitopsQueue chan *implementation.Gitops `json:"-"`
	Ctx         context.Context             `json:"-"`
	Cancel      context.CancelFunc          `json:"-"`
	Ticker      *time.Ticker                `json:"-"`
	Logger      *zap.Logger                 `json:"-"`
	LogPath     string                      `json:"-"`
}

type BackOff struct {
	BackOff bool
	Failure int
}
