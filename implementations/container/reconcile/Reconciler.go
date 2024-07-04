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
	"go.uber.org/zap"
	"time"
)

func NewWatcher(containerObj *container.Container) *watcher.Container {
	interval := 5 * time.Second
	ctx, fn := context.WithCancel(context.Background())

	return &watcher.Container{
		Container:      containerObj,
		Syncing:        false,
		Tracking:       false,
		ContainerQueue: make(chan *container.Container),
		Ctx:            ctx,
		Cancel:         fn,
		Ticker:         time.NewTicker(interval),
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
		logger.Log.Info("container already reconciling, waiting for the free slot")
		return
	}

	switch containerObj.Status.GetState() {
	case status.STATUS_CREATED:
		containerObj.Status.TransitionState(status.STATUS_DEPENDS_SOLVING)
		go dependency.Ready(shared, containerObj.Static.Group, containerObj.Static.GeneratedName, containerObj.Static.Definition.Spec.Container.Dependencies)

		shared.Watcher.Find(fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)).ContainerQueue <- containerObj

		break
	case status.STATUS_DEPENDS_SOLVING:
		logger.Log.Info("Solving dependencies for the container", zap.String("container", containerObj.Static.GeneratedName))

		shared.Watcher.Find(fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)).ContainerQueue <- containerObj
		break
	case status.STATUS_DEPENDS_SOLVED:
		logger.Log.Info("trying to run container", zap.String("group", containerObj.Static.Group), zap.String("name", containerObj.Static.Name))

		// Fix GitOps reconcile!!!!!!!!
		//container.SetOwner(c.Request.Header.Get("Owner"))
		containerObj.Prepare(shared.Client)
		_, err := containerObj.Run(shared.Manager.Config.Environment, shared.Client, shared.DnsCache)

		if err == nil {
			containerObj.Status.TransitionState(status.STATUS_RUNNING)

			go containerObj.Ready(shared.Client, err)
		} else {
			containerObj.Status.TransitionState(status.STATUS_DEAD)
		}

		shared.Watcher.Find(fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)).ContainerQueue <- containerObj

		break
	case status.STATUS_DEPENDS_FAILED:
		logger.Log.Info("Status is depends failed 4eva")
		break
	case status.STATUS_READINESS:
		logger.Log.Info("Solving readiness for the container", zap.String("container", containerObj.Static.GeneratedName))
		break
	case status.STATUS_READINESS_FAILED:
		containerObj.Stop()
		WaitForStop(containerObj)

		shared.Watcher.Find(fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)).ContainerQueue <- containerObj

		break
	case status.STATUS_KILLED:
		// NOOP: wait for dead
		break
	case status.STATUS_READY:
		containerObj.Status.TransitionState(status.STATUS_RUNNING)
		shared.Watcher.Find(fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)).ContainerQueue <- containerObj

		break
	case status.STATUS_DRIFTED:
		logger.Log.Info("sending container to reconcile state", zap.String("container", containerObj.Static.GeneratedName))
		shared.Watcher.Find(fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)).ContainerQueue <- containerObj
		break
	case status.STATUS_DEAD:
		shared.Registry.BackOffTracking(containerObj.Static.Group, containerObj.Static.GeneratedName)

		if shared.Registry.BackOffTracker[containerObj.Static.Group][containerObj.Static.GeneratedName] > 5 {
			logger.Log.Error(fmt.Sprintf("%s container is backoff restarting", containerObj.Static.GeneratedName))

			shared.Registry.BackOffReset(containerObj.Static.Group, containerObj.Static.GeneratedName)

			containerObj.Status.TransitionState(status.STATUS_BACKOFF)
		} else {
			containerObj.Delete()
			containerObj.Status.TransitionState(status.STATUS_CREATED)
		}

		shared.Watcher.Find(fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)).ContainerQueue <- containerObj

		break
	case status.STATUS_BACKOFF:
		logger.Log.Info(fmt.Sprintf("%s container is in backoff state", containerObj.Static.GeneratedName))
		break
	case status.STATUS_RUNNING:
		// NOOP
		break
	case status.STATUS_PENDING_DELETE:
		shared.Registry.Remove(containerObj.Static.Group, containerObj.Static.GeneratedName)
		containerObj.Stop()

		WaitForStop(containerObj)

		err := containerObj.Delete()

		if err != nil {
		}

		containerWatcher.Cancel()

		break
	}

	containerObj.Status.Reconciling = false
}
