package reconcile

import (
	"context"
	"fmt"
	"github.com/simplecontainer/smr/implementations/container/container"
	"github.com/simplecontainer/smr/implementations/container/dependency"
	"github.com/simplecontainer/smr/implementations/container/shared"
	"github.com/simplecontainer/smr/implementations/container/status"
	"github.com/simplecontainer/smr/implementations/container/watcher"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/manager"
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

	return &watcher.Container{
		Container:      containerObj,
		Syncing:        false,
		Tracking:       false,
		ContainerQueue: make(chan *container.Container),
		Ctx:            ctx,
		Cancel:         fn,
		Ticker:         time.NewTicker(interval),
		Logger:         loggerObj,
	}
}

func HandleTickerAndEvents(shared *shared.Shared, containerWatcher *watcher.Container) {
	for {
		select {
		case <-containerWatcher.Ctx.Done():
			containerWatcher.Ticker.Stop()
			close(containerWatcher.ContainerQueue)
			shared.Watcher.Remove(fmt.Sprintf("%s.%s", containerWatcher.Container.Static.Group, containerWatcher.Container.Static.Name))

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

		shared.Watcher.Find(fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)).ContainerQueue <- containerObj

		break
	case status.STATUS_DEPENDS_SOLVING:
		containerWatcher.Logger.Info("Solving dependencies for the container", zap.String("container", containerObj.Static.GeneratedName))
		shared.Watcher.Find(fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)).ContainerQueue <- containerObj
		break
	case status.STATUS_DEPENDS_SOLVED:
		containerWatcher.Logger.Info("trying to run container", zap.String("group", containerObj.Static.Group), zap.String("name", containerObj.Static.Name))

		// Fix GitOps reconcile!!!!!!!!
		//container.SetOwner(c.Request.Header.Get("Owner"))
		err := containerObj.Prepare(shared.Client)

		if err == nil {
			_, err = containerObj.Run(shared.Manager.Config.Environment, shared.Client, shared.DnsCache)

			if err == nil {
				containerObj.Status.TransitionState(containerObj.Static.GeneratedName, status.STATUS_RUNNING)

				go containerObj.Ready(shared.Client)
			} else {
				containerObj.Status.TransitionState(containerObj.Static.GeneratedName, status.STATUS_DEAD)
			}
		} else {
			containerObj.Status.TransitionState(containerObj.Static.GeneratedName, status.STATUS_INVALID_CONFIGURATION)
		}

		shared.Watcher.Find(fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)).ContainerQueue <- containerObj
		break
	case status.STATUS_DEPENDS_FAILED:
		containerWatcher.Logger.Info("Status is depends failed 4eva")
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

		containerWatcher.Logger.Info("Solving readiness for the container", zap.String("container", containerObj.Static.GeneratedName))
		break
	case status.STATUS_READINESS_FAILED:
		c, err := containerObj.Get()

		if err != nil {
			logger.Log.Info(err.Error())
			break
		}

		if c.State == "running" {
			containerWatcher.Logger.Info("stopping container", zap.String("container", containerObj.Static.GeneratedName))

			containerObj.Stop()
			WaitForStop(containerObj)
		}

		containerObj.Status.TransitionState(containerObj.Static.GeneratedName, status.STATUS_DEAD)
		shared.Watcher.Find(fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)).ContainerQueue <- containerObj

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
		shared.Watcher.Find(fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)).ContainerQueue <- containerObj

		break
	case status.STATUS_DRIFTED:
		containerWatcher.Logger.Info("sending container to reconcile state", zap.String("container", containerObj.Static.GeneratedName))
		shared.Watcher.Find(fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)).ContainerQueue <- containerObj
		break
	case status.STATUS_DEAD:
		c, err := containerObj.Get()

		if err != nil {
			logger.Log.Info(err.Error())
			break
		}

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
			containerWatcher.Logger.Info("waiting to die",
				zap.String("container", containerObj.Static.GeneratedName),
				zap.String("current-state", c.State),
			)
		}

		shared.Watcher.Find(fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)).ContainerQueue <- containerObj

		break
	case status.STATUS_BACKOFF:
		containerWatcher.Logger.Info(fmt.Sprintf("%s container is in backoff state", containerObj.Static.GeneratedName))
		break
	case status.STATUS_INVALID_CONFIGURATION:
		break
	case status.STATUS_RUNNING:
		// NOOP
		break
	case status.STATUS_PENDING_DELETE:
		c, err := containerObj.Get()

		if err != nil {
			shared.Registry.Remove(containerObj.Static.Group, containerObj.Static.GeneratedName)
			containerWatcher.Cancel()
		} else {
			if c.State == "running" {
				for _, readinessElem := range containerObj.Static.Readiness {
					readinessElem.Cancel()
				}

				containerObj.Stop()
				WaitForStop(containerObj)
			}

			err = containerObj.Delete()

			if err != nil {
				containerWatcher.Logger.Info("failed to delete container from docker daemon",
					zap.String("container", containerObj.Static.GeneratedName),
				)
			} else {
				shared.Registry.Remove(containerObj.Static.Group, containerObj.Static.GeneratedName)
				containerWatcher.Cancel()
			}
		}
		break
	}

	containerObj.Status.Reconciling = false
}
