package reconcile

import (
	"context"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/simplecontainer/smr/pkg/authentication"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/dependency"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/readiness"
	"github.com/simplecontainer/smr/pkg/kinds/container/shared"
	"github.com/simplecontainer/smr/pkg/kinds/container/status"
	"github.com/simplecontainer/smr/pkg/kinds/container/watcher"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/static"
	"go.uber.org/zap"
	"os"
	"time"
)

func NewWatcher(containerObj platforms.IContainer, mgr *manager.Manager, user *authentication.User) *watcher.Container {
	interval := 5 * time.Second
	ctx, fn := context.WithCancel(context.Background())

	loggerObj := logger.NewLogger(os.Getenv("LOG_LEVEL"), []string{fmt.Sprintf("/tmp/container.%s.%s.log", containerObj.GetGroup(), containerObj.GetGeneratedName())}, []string{fmt.Sprintf("/tmp/container.%s.%s.log", containerObj.GetGroup(), containerObj.GetGeneratedName())})

	return &watcher.Container{
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

func HandleTickerAndEvents(shared *shared.Shared, containerWatcher *watcher.Container) {
	for {
		select {
		case <-containerWatcher.Ctx.Done():
			containerWatcher.Ticker.Stop()

			close(containerWatcher.ContainerQueue)
			close(containerWatcher.ReadinessChan)
			close(containerWatcher.DependencyChan)

			shared.Registry.Remove(containerWatcher.Container.GetGroup(), containerWatcher.Container.GetGeneratedName())
			shared.Watcher.Remove(containerWatcher.Container.GetGroupIdentifier())

			return
		case <-containerWatcher.ContainerQueue:
			containerWatcher.Ticker.Reset(5 * time.Second)
			go Container(shared, containerWatcher)
			break
		case <-containerWatcher.Ticker.C:
			if containerWatcher.Container.GetStatus().GetCategory() != status.CATEGORY_END {
				go Container(shared, containerWatcher)
			} else {
				containerWatcher.Ticker.Stop()
			}
			break
		}
	}
}

func Container(shared *shared.Shared, containerWatcher *watcher.Container) {
	containerObj := containerWatcher.Container

	if containerObj.GetStatus().Reconciling {
		containerWatcher.Logger.Info("container already reconciling, waiting for the free slot")
		return
	}

	containerObj.GetStatus().Reconciling = true

	switch containerObj.GetStatus().GetState() {
	case status.STATUS_CREATED:
		containerState := GetState(containerWatcher)
		containerObj.GetStatus().Recreated = false

		switch containerState {
		case "created":
			// No OP do check again
			break
		case "exited":
			containerWatcher.Logger.Info("container created but already exited")
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_DEAD)
			break
		case "dead":
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_KILL)
			containerWatcher.Logger.Info("container died while readiness checking")
			break
		case "removing":
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_KILL)
			containerWatcher.Logger.Info("container died while readiness checking")
			break
		case "removed":
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_KILL)
			containerWatcher.Logger.Info("container died while readiness checking")
			break
		case "running":
			containerWatcher.Logger.Info("container created but already running")
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_KILL)
			break
		default:
			containerWatcher.Logger.Info("container created")
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_PREPARE)
		}

		ReconcileLoop(containerWatcher)
		break
	case status.STATUS_RECREATED:
		containerState := GetState(containerWatcher)

		switch containerState {
		case "created":
			// No OP do check again
			break
		case "exited":
			containerWatcher.Logger.Info("container created but already exited")
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_DEAD)
			break
		case "dead":
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_KILL)
			containerWatcher.Logger.Info("container died while readiness checking")
			break
		case "removing":
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_KILL)
			containerWatcher.Logger.Info("container died while readiness checking")
			break
		case "removed":
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_KILL)
			containerWatcher.Logger.Info("container died while readiness checking")
			break
		case "running":
			containerObj.GetStatus().Recreated = true
			containerWatcher.Logger.Info("container recreated but already running - next restart of container will pickup changes")
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_RUNNING)
			break
		default:
			containerWatcher.Logger.Info("container recreated")
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_KILL)
		}

		ReconcileLoop(containerWatcher)
		break
	case status.STATUS_PREPARE:
		err := containerObj.Prepare(shared.Client, containerWatcher.User)

		if err == nil {
			go dependency.Ready(shared.Registry, containerObj.GetGroup(), containerObj.GetGeneratedName(), containerObj.GetDefinition().(*v1.ContainerDefinition).Spec.Container.Dependencies, containerWatcher.DependencyChan)

			containerWatcher.Logger.Info("container prepared")
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_DEPENDS_CHECKING)
		} else {
			containerWatcher.Logger.Info(err.Error())
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_PENDING)
		}

		ReconcileLoop(containerWatcher)
		break
	case status.STATUS_DEPENDS_CHECKING:
		ContinueReconciliation := false
		for !ContinueReconciliation {
			select {
			case dependencyResult := <-containerWatcher.DependencyChan:
				if dependencyResult == nil {
					ContinueReconciliation = true
					break
				}

				switch dependencyResult.State {
				case dependency.CHECKING:
					containerWatcher.Logger.Info("checking dependency")

					if dependencyResult.Error != nil {
						containerWatcher.Logger.Info(dependencyResult.Error.Error())
					}

					break
				case dependency.SUCCESS:
					containerWatcher.Logger.Info("dependency check success")
					containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_DEPENDS_SOLVED)

					ContinueReconciliation = true
					break
				case dependency.FAILED:
					containerWatcher.Logger.Info("dependency check failed")
					containerWatcher.Logger.Info(dependencyResult.Error.Error())
					containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_DEPENDS_FAILED)

					ContinueReconciliation = true
					break
				}
			}
		}

		ReconcileLoop(containerWatcher)
		break
	case status.STATUS_DEPENDS_SOLVED:
		containerObj.GetStatus().LastDependsSolved = true
		containerObj.GetStatus().LastDependsSolvedTimestamp = time.Now()

		containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_START)

		ReconcileLoop(containerWatcher)
		break
	case status.STATUS_START:
		containerState := GetState(containerWatcher)
		containerObj.GetStatus().Recreated = false
		var err error = nil

		switch containerState {
		case "created":
			// No OP do check again
			break
		case "exited":
			containerWatcher.Logger.Info("container created but already exited")
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_DEAD)
			break
		case "dead":
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_KILL)
			containerWatcher.Logger.Info("container died while readiness checking")
			break
		case "removing":
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_KILL)
			containerWatcher.Logger.Info("container died while readiness checking")
			break
		case "removed":
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_KILL)
			containerWatcher.Logger.Info("container died while readiness checking")
			break
		case "running":
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_KILL)
			containerWatcher.Logger.Info("container started while running transition to kill")
			break
		default:
			containerWatcher.Logger.Info("container attempt to start")
			_, err = containerObj.Run(shared.Manager.Config, shared.Client, shared.DnsCache, containerWatcher.User)

			if err == nil {
				containerWatcher.Logger.Info("container started")
				containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_READINESS_CHECKING)

				go func() {
					_, err = readiness.Ready(shared.Client, containerObj, containerWatcher.User, containerWatcher.ReadinessChan, containerWatcher.Logger)
					if err != nil {
						containerWatcher.Logger.Error(err.Error())
					}
				}()
			} else {
				containerWatcher.Logger.Info("container start failed", zap.Error(err))
				containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_RUNTIME_PENDING)
			}
		}

		ReconcileLoop(containerWatcher)
		break
	case status.STATUS_READINESS_CHECKING:
		containerState := GetState(containerWatcher)

		switch containerState {
		case "created":
			// No OP do check again
			break
		case "exited":
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_KILL)
			containerWatcher.Logger.Info("container died while readiness checking")
			break
		case "dead":
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_KILL)
			containerWatcher.Logger.Info("container died while readiness checking")
			break
		case "removing":
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_KILL)
			containerWatcher.Logger.Info("container died while readiness checking")
			break
		case "removed":
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_KILL)
			containerWatcher.Logger.Info("container died while readiness checking")
			break
		case "running":
			ContinueReconciliation := false
			for !ContinueReconciliation {
				select {
				case readinessResult := <-containerWatcher.ReadinessChan:
					containerState = GetState(containerWatcher)

					if containerState == "running" {
						switch readinessResult.State {
						case dependency.CHECKING:
							containerWatcher.Logger.Info("checking readiness")
							break
						case dependency.SUCCESS:
							containerWatcher.Logger.Info("readiness check success")
							containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_READY)

							ContinueReconciliation = true
							break
						case dependency.FAILED:
							containerWatcher.Logger.Info("readiness check failed")
							containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_READINESS_FAILED)

							ContinueReconciliation = true
							break
						}
					}
				}
			}
			break
		}

		ReconcileLoop(containerWatcher)
		break
	case status.STATUS_READY:
		containerObj.GetStatus().LastReadiness = true
		containerObj.GetStatus().LastReadinessTimestamp = time.Now()

		containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_RUNNING)
		ReconcileLoop(containerWatcher)

		break
	case status.STATUS_RUNNING:
		containerStateEngine, err := containerObj.GetContainerState()

		if err != nil {
			ReconcileLoop(containerWatcher)
			break
		}

		switch containerStateEngine {
		case "created":
			// No OP do check again
			break
		case "exited":
			containerWatcher.Logger.Info("container is dead")
			shared.Registry.BackOffTracking(containerObj.GetGroup(), containerObj.GetGeneratedName())

			if shared.Registry.BackOffTracker[containerObj.GetGroup()][containerObj.GetGeneratedName()] > 5 {
				containerWatcher.Logger.Error(fmt.Sprintf("%s container is backoff restarting", containerObj.GetGeneratedName()))

				shared.Registry.BackOffReset(containerObj.GetGroup(), containerObj.GetGeneratedName())

				containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_BACKOFF)
			} else {
				containerObj.Delete()
				containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_PREPARE)
			}
			break
		case "dead":
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_KILL)
			containerWatcher.Logger.Info("container died while readiness checking")
			break
		case "removing":
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_KILL)
			containerWatcher.Logger.Info("container died while readiness checking")
			break
		case "removed":
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_KILL)
			containerWatcher.Logger.Info("container died while readiness checking")
			break
		case "running":
			containerWatcher.Logger.Debug("container is running")
			containerObj.GetStatus().Reconciling = false
			return
		}

		ReconcileLoop(containerWatcher)
		break
	case status.STATUS_DEAD:
		containerStateEngine, err := containerObj.GetContainerState()

		if err != nil {
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_PREPARE)
			ReconcileLoop(containerWatcher)

			return
		}

		switch containerStateEngine {
		case "created":
			containerWatcher.Logger.Info("container couldn't be created")
			shared.Registry.BackOffTracking(containerObj.GetGroup(), containerObj.GetGeneratedName())

			containerObj.Delete()
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_PREPARE)
			ReconcileLoop(containerWatcher)
			break
		case "exited":
			containerWatcher.Logger.Info("container is dead")
			shared.Registry.BackOffTracking(containerObj.GetGroup(), containerObj.GetGeneratedName())

			if shared.Registry.BackOffTracker[containerObj.GetGroup()][containerObj.GetGeneratedName()] > 5 {
				containerWatcher.Logger.Error(fmt.Sprintf("%s container is backoff restarting", containerObj.GetGeneratedName()))

				shared.Registry.BackOffReset(containerObj.GetGroup(), containerObj.GetGeneratedName())

				containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_BACKOFF)
				ReconcileLoop(containerWatcher)
			} else {
				containerWatcher.Logger.Info("deleting dead container")

				err = containerObj.Delete()
				if err != nil {
					containerWatcher.Logger.Error(err.Error())
				}

				containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_PREPARE)
				ReconcileLoop(containerWatcher)
			}
			break
		case "dead":
			containerWatcher.Logger.Info("container dead - cleanup", zap.String("current-state", containerStateEngine))
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_KILL)
			ReconcileLoop(containerWatcher)

			break
		case "removing":
			// No OP
			break
		case "removed":
			containerWatcher.Logger.Info("container removed - cleanup", zap.String("current-state", containerStateEngine))
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_KILL)
			ReconcileLoop(containerWatcher)
			break
		}

		break
	case status.STATUS_BACKOFF:
		containerWatcher.Logger.Info("container is in backoff state")
		break
	case status.STATUS_DEPENDS_FAILED:
		containerWatcher.Logger.Info("container depends failed")
		containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_PREPARE)

		ReconcileLoop(containerWatcher)
		break
	case status.STATUS_READINESS_FAILED:
		containerWatcher.Logger.Info("container readiness failed")
		containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_KILL)

		ReconcileLoop(containerWatcher)
		break
	case status.STATUS_PENDING:
		containerWatcher.Logger.Info("container invalid configuration")
		err := containerObj.Prepare(shared.Client, containerWatcher.User)

		if err == nil {
			go dependency.Ready(shared.Registry, containerObj.GetGroup(), containerObj.GetGeneratedName(), containerObj.GetDefinition().(*v1.ContainerDefinition).Spec.Container.Dependencies, containerWatcher.DependencyChan)

			containerWatcher.Logger.Info("container prepared")
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_DEPENDS_CHECKING)
			ReconcileLoop(containerWatcher)
		} else {
			containerWatcher.Logger.Info(err.Error())
			containerWatcher.Container.GetStatus().Reconciling = false
		}
		break

	case status.STATUS_RUNTIME_PENDING:
		containerWatcher.Logger.Info("container engine runtime returned error - will retry till conditions met")
		containerWatcher.Retry += 1

		if containerWatcher.Retry > 12 {
			containerWatcher.Retry = 0
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_KILL)
		} else {
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_PREPARE)
		}

		containerWatcher.Container.GetStatus().Reconciling = false
		break

	case status.STATUS_KILL:
		containerStateEngine, err := containerObj.GetContainerState()

		if err != nil {
			containerWatcher.Logger.Info(err.Error())
			containerWatcher.Container.GetStatus().Reconciling = false
			break
		}

		switch containerStateEngine {
		case "created":
			err = containerObj.Delete()

			if err != nil {
				containerWatcher.Logger.Info("failed to delete container from engine daemon")
				containerWatcher.Logger.Error(err.Error())
			} else {
				containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_DEAD)
			}
			break
		case "exited":
			err = containerObj.Delete()

			if err != nil {
				containerWatcher.Logger.Info("failed to delete container from engine daemon")
				containerWatcher.Logger.Error(err.Error())
			} else {
				containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_DEAD)
			}
			break
		case "dead":
			err = containerObj.Delete()

			if err != nil {
				containerWatcher.Logger.Info("failed to delete container from engine daemon")
				containerWatcher.Logger.Error(err.Error())
			} else {
				containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_DEAD)
			}
			break
		case "removing":
			// No OP
			break
		case "removed":
			err = containerObj.Delete()

			if err != nil {
				containerWatcher.Logger.Info("failed to delete container from engine daemon")
				containerWatcher.Logger.Error(err.Error())
			} else {
				containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_DEAD)
			}
			break
		default:
			containerWatcher.Logger.Info("attempt to shutdown gracefully")

			err = containerObj.Stop(static.SIGTERM)

			if err != nil {
				err = containerObj.Stop(static.SIGKILL)

				if err != nil {
					containerWatcher.Logger.Error(err.Error())
				}
			}
		}

		ReconcileLoop(containerWatcher)
		break
	case status.STATUS_PENDING_DELETE:
		containerStateEngine, err := containerObj.GetContainerState()

		if err != nil {
			containerWatcher.Logger.Info(err.Error())
			containerWatcher.Container.GetStatus().Reconciling = false
			break
		}

		switch containerStateEngine {
		case "created":
			err = containerObj.Delete()

			if err != nil {
				containerWatcher.Logger.Info("failed to delete container from engine daemon")
				containerWatcher.Logger.Error(err.Error())
			} else {
				containerWatcher.Cancel()
			}
			break
		case "exited":
			err = containerObj.Delete()

			if err != nil {
				containerWatcher.Logger.Info("failed to delete container from engine daemon")
				containerWatcher.Logger.Error(err.Error())
			} else {
				containerWatcher.Cancel()
			}
			break
		case "dead":
			err = containerObj.Delete()

			if err != nil {
				containerWatcher.Logger.Info("failed to delete container from engine daemon")
				containerWatcher.Logger.Error(err.Error())
			} else {
				containerWatcher.Cancel()
			}
			break
		case "removing":
			// No OP
			break
		case "removed":
			err = containerObj.Delete()

			if err != nil {
				containerWatcher.Logger.Info("failed to delete container from engine daemon")
				containerWatcher.Logger.Error(err.Error())
			} else {
				containerWatcher.Cancel()
			}
			break
		case "running":
			err = containerObj.Stop(static.SIGTERM)

			if err != nil {
				err = containerObj.Stop(static.SIGKILL)

				if err != nil {
					containerWatcher.Logger.Error(err.Error())
				}
			}
			break
		}
		break
	}
}

func ReconcileLoop(containerWatcher *watcher.Container) {
	containerWatcher.Container.GetStatus().Reconciling = false
	containerWatcher.ContainerQueue <- containerWatcher.Container
}

func GetState(containerWatcher *watcher.Container) string {
	containerStateEngine, err := containerWatcher.Container.GetContainerState()

	if err != nil {
		return ""
	}

	return containerStateEngine
}

func Wait(timeout time.Duration, f func() error) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	backOff := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)

	return backoff.Retry(f, backOff)
}
