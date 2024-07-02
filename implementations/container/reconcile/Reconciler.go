package reconcile

import (
	"context"
	"fmt"
	"github.com/qdnqn/smr/pkg/container"
	"github.com/qdnqn/smr/pkg/dependency"
	"github.com/qdnqn/smr/pkg/logger"
	"github.com/qdnqn/smr/pkg/manager"
	"github.com/qdnqn/smr/pkg/reconciler"
	"github.com/qdnqn/smr/pkg/status"
	"go.uber.org/zap"
	"time"
)

func NewWatcher(containerObj *container.Container) *reconciler.Container {
	interval := 5 * time.Second
	ctx, fn := context.WithCancel(context.Background())

	return &reconciler.Container{
		Container:      containerObj,
		Syncing:        false,
		Tracking:       false,
		ContainerQueue: make(chan *container.Container),
		Ctx:            ctx,
		Cancel:         fn,
		Ticker:         time.NewTicker(interval),
	}
}

func HandleTickerAndEvents(mgr *manager.Manager, containerWatcher *reconciler.Container) {
	for {
		select {
		case <-containerWatcher.Ctx.Done():
			containerWatcher.Ticker.Stop()
			close(containerWatcher.ContainerQueue)
			mgr.ContainerWatchers.Remove(fmt.Sprintf("%s.%s", containerWatcher.Container.Static.Group, containerWatcher.Container.Static.Name))

			return
		case <-containerWatcher.ContainerQueue:
			go ReconcileContainer(mgr, containerWatcher)
			break
		case <-containerWatcher.Ticker.C:
			fmt.Println("TICK - CONTAINER")
			if !containerWatcher.Container.Status.Reconciling {
				go ReconcileContainer(mgr, containerWatcher)
			}
			break
		}
	}
}

func ReconcileContainer(mgr *manager.Manager, containerWatcher *reconciler.Container) {
	containerObj := containerWatcher.Container

	if containerObj.Status.Reconciling {
		logger.Log.Info("container already reconciling, waiting for the free slot")
		return
	}

	switch containerObj.Status.GetState() {
	case status.STATUS_CREATED:
		containerObj.Status.TransitionState(status.STATUS_DEPENDS_SOLVING)
		go dependency.Ready(mgr, containerObj.Static.Group, containerObj.Static.GeneratedName, containerObj.Static.Definition.Spec.Container.Dependencies)

		mgr.ContainerWatchers.Find(fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)).ContainerQueue <- containerObj

		break
	case status.STATUS_DEPENDS_SOLVING:
		logger.Log.Info("Solving dependencies for the container", zap.String("container", containerObj.Static.GeneratedName))

		mgr.ContainerWatchers.Find(fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)).ContainerQueue <- containerObj
		break
	case status.STATUS_DEPENDS_SOLVED:
		logger.Log.Info("trying to run container", zap.String("group", containerObj.Static.Group), zap.String("name", containerObj.Static.Name))

		// Fix GitOps reconcile!!!!!!!!
		//container.SetOwner(c.Request.Header.Get("Owner"))
		containerObj.Prepare(mgr.Badger)
		_, err := containerObj.Run(mgr.Runtime, mgr.Badger, mgr.BadgerEncrypted, mgr.DnsCache)

		if err == nil {
			containerObj.Status.TransitionState(status.STATUS_RUNNING)

			client, err := mgr.Keys.GenerateHttpClient()
			go containerObj.Ready(mgr.BadgerEncrypted, client, err)
		} else {
			containerObj.Status.TransitionState(status.STATUS_DEAD)
		}

		mgr.ContainerWatchers.Find(fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)).ContainerQueue <- containerObj

		break
	case status.STATUS_DEPENDS_FAILED:
		logger.Log.Info("Status is depends failed 4eva")
		break
	case status.STATUS_READINESS:
		logger.Log.Info("Solving readiness for the container", zap.String("container", containerObj.Static.GeneratedName))
		break
	case status.STATUS_READINESS_FAILED:
		containerObj.Status.TransitionState(status.STATUS_DEAD)
		mgr.ContainerWatchers.Find(fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)).ContainerQueue <- containerObj

		break
	case status.STATUS_READY:
		containerObj.Status.TransitionState(status.STATUS_RUNNING)
		mgr.ContainerWatchers.Find(fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)).ContainerQueue <- containerObj

		break
	case status.STATUS_DRIFTED:
		logger.Log.Info("sending container to reconcile state", zap.String("container", containerObj.Static.GeneratedName))
		mgr.ContainerWatchers.Find(fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)).ContainerQueue <- containerObj
		break
	case status.STATUS_DEAD:
		mgr.Registry.BackOffTracking(containerObj.Static.Group, containerObj.Static.GeneratedName)

		if mgr.Registry.BackOffTracker[containerObj.Static.Group][containerObj.Static.GeneratedName] > 5 {
			logger.Log.Error(fmt.Sprintf("%s container is backoff restarting", containerObj.Static.GeneratedName))

			mgr.Registry.BackOffReset(containerObj.Static.Group, containerObj.Static.GeneratedName)

			containerObj.Status.TransitionState(status.STATUS_BACKOFF)
		} else {
			containerObj.Delete()
			containerObj.Status.TransitionState(status.STATUS_CREATED)
		}

		mgr.ContainerWatchers.Find(fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)).ContainerQueue <- containerObj

		break
	case status.STATUS_BACKOFF:
		logger.Log.Info(fmt.Sprintf("%s container is in backoff state", containerObj.Static.GeneratedName))
		break
	case status.STATUS_RUNNING:
		// NOOP
		break
	case status.STATUS_PENDING_DELETE:
		mgr.Registry.Remove(containerObj.Static.Group, containerObj.Static.GeneratedName)
		containerObj.Stop()

		timeout := false
		waitForStop := make(chan string, 1)
		go func() {
			for {
				c := containerObj.Get()

				if timeout {
					return
				}

				if c != nil && c.State != "exited" {
					logger.Log.Info(fmt.Sprintf("waiting for container to exit %s", containerObj.Static.GeneratedName))
					time.Sleep(1 * time.Second)
				} else {
					break
				}
			}

			waitForStop <- "container exited proceed with delete for reconciliation"
		}()

		select {
		case res := <-waitForStop:
			logger.Log.Info(fmt.Sprintf("%s %s", res, containerObj.Static.GeneratedName))
		case <-time.After(30 * time.Second):
			logger.Log.Info("timed out waiting for the container to exit", zap.String("container", containerObj.Static.GeneratedName))
			timeout = true
		}

		err := containerObj.Delete()

		if err != nil {
		}

		containerWatcher.Cancel()

		break
	}

	containerObj.Status.Reconciling = false
}
