package reconcile

import (
	"context"
	"fmt"
	"github.com/simplecontainer/smr/implementations/container/container"
	"github.com/simplecontainer/smr/implementations/container/dependency"
	"github.com/simplecontainer/smr/implementations/container/shared"
	"github.com/simplecontainer/smr/implementations/container/status"
	"github.com/simplecontainer/smr/implementations/container/watcher"
	hubShared "github.com/simplecontainer/smr/implementations/hub/shared"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/plugins"
	"go.uber.org/zap"
	"time"
)

func NewWatcher(containerObj *container.Container, mgr *manager.Manager) *watcher.Container {
	interval := 5 * time.Second
	ctx, fn := context.WithCancel(context.Background())

	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{fmt.Sprintf("/tmp/container.%s.%s.log", containerObj.Static.Group, containerObj.Static.GeneratedName)}

	loggerObj, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	pl := plugins.GetPlugin(mgr.Config.Root, "hub.so")
	sharedContainer := pl.GetShared().(*hubShared.Shared)

	return &watcher.Container{
		Container:      containerObj,
		Syncing:        false,
		ContainerQueue: make(chan *container.Container),
		Ctx:            ctx,
		Cancel:         fn,
		Ticker:         time.NewTicker(interval),
		Logger:         loggerObj,
		EventChannel:   sharedContainer.Event,
	}
}

func HandleTickerAndEvents(shared *shared.Shared, containerWatcher *watcher.Container) {
	for {
		select {
		case <-containerWatcher.Ctx.Done():
			containerWatcher.Ticker.Stop()

			close(containerWatcher.ContainerQueue)
			shared.Watcher.Remove(containerWatcher.Container.GetGroupIdentifier())

			return
		case <-containerWatcher.ContainerQueue:
			go ReconcileContainer(shared, containerWatcher)
			break
		case <-containerWatcher.Ticker.C:
			if !containerWatcher.Container.Status.Reconciling {
				go ReconcileContainer(shared, containerWatcher)
			}
			break
		}
	}
}

func ReconcileContainer(shared *shared.Shared, containerWatcher *watcher.Container) {
	containerObj := containerWatcher.Container

	if containerObj.Status.Reconciling {
		containerWatcher.Logger.Info("container already reconciling, waiting for the free slot")
		return
	}

	containerObj.Status.Reconciling = true

	switch containerObj.Status.GetState() {
	case status.STATUS_CREATED:
		containerObj.Status.TransitionState(containerObj.Static.GeneratedName, status.STATUS_DEPENDS_SOLVING)
		go dependency.Ready(shared, containerObj.Static.Group, containerObj.Static.GeneratedName, containerObj.Static.Definition.Spec.Container.Dependencies)

		containerWatcher.ContainerQueue <- containerObj

		break
	case status.STATUS_DEPENDS_SOLVING:
		containerWatcher.Logger.Info("solving dependencies for the container")
		containerWatcher.ContainerQueue <- containerObj
		break
	case status.STATUS_DEPENDS_SOLVED:
		containerWatcher.Logger.Info("prepare container attempt")

		// Fix GitOps reconcile!!!!!!!!
		//container.SetOwner(c.Request.Header.Get("Owner"))
		err := containerObj.Prepare(shared.Client)

		if err == nil {
			containerWatcher.Logger.Info("container prepared")
			containerWatcher.Logger.Info("container run attempt")

			_, err = containerObj.Run(shared.Manager.Config.Environment, shared.Client, shared.DnsCache)

			if err == nil {
				containerObj.Status.TransitionState(containerObj.Static.GeneratedName, status.STATUS_RUNNING)
				containerWatcher.Logger.Info("container running")

				go containerObj.Ready(shared.Client)
			} else {
				containerWatcher.Logger.Info("container running failed", zap.Error(err))
				containerObj.Status.TransitionState(containerObj.Static.GeneratedName, status.STATUS_DEAD)
			}
		} else {
			containerWatcher.Logger.Info("container prepare failed", zap.Error(err))
			containerObj.Status.TransitionState(containerObj.Static.GeneratedName, status.STATUS_INVALID_CONFIGURATION)
		}

		containerWatcher.ContainerQueue <- containerObj
		break
	case status.STATUS_DEPENDS_FAILED:
		containerWatcher.Logger.Info("dependency check failed")
		break
	case status.STATUS_READINESS:
		c, err := containerObj.Get()

		if err != nil {
			logger.Log.Info(err.Error())
			break
		}

		if c.State != "running" {
			containerObj.Status.TransitionState(containerObj.Static.GeneratedName, status.STATUS_DEAD)
		}

		containerWatcher.Logger.Info("solving readiness checks")
		break
	case status.STATUS_READINESS_FAILED:
		c, err := containerObj.Get()

		if err != nil {
			logger.Log.Info(err.Error())
			break
		}

		if c.State == "running" {
			containerWatcher.Logger.Info("stopping container")

			containerObj.Stop()
			WaitForStop(containerObj)
		}

		containerObj.Status.TransitionState(containerObj.Static.GeneratedName, status.STATUS_DEAD)
		containerWatcher.ContainerQueue <- containerObj

		break
	case status.STATUS_KILLED:
		// NOOP: wait for dead
		break
	case status.STATUS_READY:
		c, err := containerObj.Get()

		if err != nil {
			logger.Log.Info(err.Error())
			break
		}

		if c.State != "running" {
			containerObj.Status.TransitionState(containerObj.Static.GeneratedName, status.STATUS_DEAD)
		} else {
			containerObj.Status.TransitionState(containerObj.Static.GeneratedName, status.STATUS_RUNNING)
		}

		containerWatcher.ContainerQueue <- containerObj

		break
	case status.STATUS_DRIFTED:
		containerWatcher.Logger.Info("sending container to reconcile state", zap.String("container", containerObj.Static.GeneratedName))
		containerWatcher.ContainerQueue <- containerObj
		break
	case status.STATUS_DEAD:
		c, err := containerObj.Get()

		if err != nil {
			containerWatcher.Logger.Info(err.Error())
			break
		}

		containerWatcher.Logger.Info("container is dead")

		if c.State == "exited" {
			shared.Registry.BackOffTracking(containerObj.Static.Group, containerObj.Static.GeneratedName)

			if shared.Registry.BackOffTracker[containerObj.Static.Group][containerObj.Static.GeneratedName] > 5 {
				containerWatcher.Logger.Error(fmt.Sprintf("%s container is backoff restarting", containerObj.Static.GeneratedName))

				shared.Registry.BackOffReset(containerObj.Static.Group, containerObj.Static.GeneratedName)

				containerObj.Status.TransitionState(containerObj.Static.GeneratedName, status.STATUS_BACKOFF)
			} else {
				containerObj.Delete()
				containerObj.Status.TransitionState(containerObj.Static.GeneratedName, status.STATUS_CREATED)
			}
		} else {
			containerWatcher.Logger.Info("waiting to die", zap.String("current-state", c.State))
		}

		containerWatcher.ContainerQueue <- containerObj

		break
	case status.STATUS_BACKOFF:
		containerWatcher.Logger.Info("container is in backoff state")
		break
	case status.STATUS_INVALID_CONFIGURATION:
		break
	case status.STATUS_RUNNING:
		// NOOP
		break
	case status.STATUS_PENDING_DELETE:
		c, err := containerObj.Get()

		containerWatcher.Logger.Info("container is pending delete")

		if err != nil {
			shared.Registry.Remove(containerObj.Static.Group, containerObj.Static.GeneratedName)
			containerWatcher.Cancel()
		} else {
			if c.State == "running" {
				containerWatcher.Logger.Info("starting graceful termination, timeout 30s")

				containerObj.Stop()
				WaitForStop(containerObj)
			}

			err = containerObj.Delete()

			if err != nil {
				containerWatcher.Logger.Info("failed to delete container from docker daemon")

			} else {
				containerWatcher.Logger.Info("container is deleted")

				shared.Registry.Remove(containerObj.Static.Group, containerObj.Static.GeneratedName)
				containerWatcher.Cancel()
			}
		}

		// Never allow reconciling to false after pending delete
		// this prevents multiple pending_delete removals
		return

		break
	}

	containerObj.Status.Reconciling = false
}
