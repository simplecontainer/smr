package watcher

import (
	"context"
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/dependency"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/readiness"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/static"
	"os"
	"strings"
	"time"
)

func New(containerObj platforms.IContainer, startState string, user *authentication.User) *Container {
	interval := 5 * time.Second
	ctx, fn := context.WithCancel(context.Background())

	format := f.New(containerObj.GetDefinition().GetPrefix(), "kind", static.KIND_CONTAINERS, containerObj.GetGroup(), containerObj.GetGeneratedName())
	path := fmt.Sprintf("/tmp/%s", strings.Replace(format.ToString(), "/", "-", -1))

	loggerObj := logger.NewLogger(os.Getenv("LOG_LEVEL"), []string{path}, []string{path})

	containerObj.GetStatus().SetState(startState)

	return &Container{
		Container:      containerObj,
		Syncing:        false,
		ContainerQueue: make(chan platforms.IContainer),
		ReadinessChan:  make(chan *readiness.ReadinessState),
		DependencyChan: make(chan *dependency.State),
		Ctx:            ctx,
		Cancel:         fn,
		Ticker:         time.NewTicker(interval),
		Retry:          0,
		Logger:         loggerObj,
		User:           user,
	}
}
