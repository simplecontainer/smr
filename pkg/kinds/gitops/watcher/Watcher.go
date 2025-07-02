package watcher

import (
	"context"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/implementation"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/manager"
	"go.uber.org/zap"
	"os"
	"time"
)

func New(gitopsObj *implementation.Gitops, mgr *manager.Manager, user *authentication.User) *Gitops {
	interval := 5 * time.Second
	ctx, fn := context.WithCancel(context.Background())

	err := logger.CreateOrRotate(gitopsObj.Gitops.LogPath)
	if err != nil {
		logger.Log.Error("failed to create log file for container", zap.String("gitops", gitopsObj.GetGroupIdentifier()))
	}

	loggerObj := logger.NewLogger(os.Getenv("LOG_LEVEL"), []string{gitopsObj.Gitops.LogPath}, []string{gitopsObj.Gitops.LogPath})

	return &Gitops{
		Gitops:      gitopsObj,
		GitopsQueue: make(chan *implementation.Gitops),
		User:        user,
		Ctx:         ctx,
		Cancel:      fn,
		Ticker:      time.NewTicker(interval),
		Poller:      time.NewTicker(gitopsObj.Gitops.PoolingInterval),
		Logger:      loggerObj,
	}
}
