package watcher

import (
	"context"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/implementation"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/manager"
	"os"
	"time"
)

func New(gitopsObj *implementation.Gitops, mgr *manager.Manager, user *authentication.User) *Gitops {
	interval := 5 * time.Second
	ctx, fn := context.WithCancel(context.Background())

	loggerObj := logger.NewLogger(os.Getenv("LOG_LEVEL"), []string{gitopsObj.LogPath}, []string{gitopsObj.LogPath})

	return &Gitops{
		Gitops:      gitopsObj,
		Children:    make([]*common.Request, 0),
		GitopsQueue: make(chan *implementation.Gitops),
		User:        user,
		Ctx:         ctx,
		Cancel:      fn,
		Ticker:      time.NewTicker(interval),
		Poller:      time.NewTicker(gitopsObj.PoolingInterval),
		Logger:      loggerObj,
	}
}
