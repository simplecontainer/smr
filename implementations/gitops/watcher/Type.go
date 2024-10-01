package watcher

import (
	"context"
	"github.com/simplecontainer/smr/implementations/gitops/gitops"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"go.uber.org/zap"
	"time"
)

type RepositoryWatcher struct {
	Repositories map[string]*Gitops
}

type Gitops struct {
	Gitops      *gitops.Gitops
	Syncing     bool
	Tracking    bool
	BackOff     BackOff
	Definition  v1.GitopsDefinition
	GitopsQueue chan *gitops.Gitops `json:"-"`
	Ctx         context.Context     `json:"-"`
	Cancel      context.CancelFunc  `json:"-"`
	Ticker      *time.Ticker        `json:"-"`
	Logger      *zap.Logger
}

type BackOff struct {
	BackOff bool
	Failure int
}
