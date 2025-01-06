package reconcile

import (
	"context"
	"errors"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/docker/docker/api/types"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/dependency"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/readiness"
	"github.com/simplecontainer/smr/pkg/kinds/container/shared"
	"github.com/simplecontainer/smr/pkg/kinds/container/status"
	"github.com/simplecontainer/smr/pkg/kinds/container/watcher"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/static"
	"go.uber.org/zap"
	"time"
)

func NewWatcher(containerObj platforms.IContainer, mgr *manager.Manager, user *authentication.User) *watcher.Container {
	interval := 5 * time.Second
	ctx, fn := context.WithCancel(context.Background())

	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{fmt.Sprintf("/tmp/container.%s.%s.log", containerObj.GetGroup(), containerObj.GetGeneratedName())}

	loggerObj, err := cfg.Build()
	if err != nil {
		panic(err)
	}

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
		dockerState := GetState(containerWatcher)

		switch dockerState {
		case "running":
			containerWatcher.Logger.Info("container created but already running")
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_KILL)
			break
		case "exited":
			containerWatcher.Logger.Info("container created but already exited")
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_DEAD)
			break
		default:
			containerWatcher.Logger.Info("container created")
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_PREPARE)
		}

		ReconcileLoop(containerWatcher)
		break
	case status.STATUS_RECREATED:
		dockerState := GetState(containerWatcher)

		switch dockerState {
		case "running":
			containerWatcher.Logger.Info("container recreated but already running")
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_PREPARE)

			break
		case "exited":
			containerWatcher.Logger.Info("container recreated but already exited")
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_DEAD)
			break
		default:
			containerWatcher.Logger.Info("container recreated")

			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_PREPARE)
		}

		ReconcileLoop(containerWatcher)
		break
	case status.STATUS_PREPARE:
		err := containerObj.Prepare(shared.Client, containerWatcher.User)

		if err == nil {
			go dependency.Ready(shared.Registry, containerObj.GetGroup(), containerObj.GetGeneratedName(), containerObj.GetDefinition().Spec.Container.Dependencies, containerWatcher.DependencyChan)

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
		dockerState := GetState(containerWatcher)
		var err error = nil

		if dockerState != "running" {
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
		} else {
			containerWatcher.Logger.Info("container is already running")

			err = containerObj.AttachToNetworks(shared.Manager.Config.Node)

			if err != nil {
				containerWatcher.Logger.Error("container and smr-agent failed to attach to all networks!")
				containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_KILL)
			} else {
				containerObj.UpdateDns(shared.DnsCache)

				containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_READINESS_CHECKING)
				go func() {
					_, err = readiness.Ready(shared.Client, containerObj, containerWatcher.User, containerWatcher.ReadinessChan, containerWatcher.Logger)

					if err != nil {
						containerWatcher.Logger.Error(err.Error())
					}
				}()
			}
		}

		ReconcileLoop(containerWatcher)
		break
	case status.STATUS_READINESS_CHECKING:
		ContinueReconciliation := false
		for !ContinueReconciliation {
			select {
			case readinessResult := <-containerWatcher.ReadinessChan:
				dockerState := GetState(containerWatcher)

				if dockerState != "running" {
					ContinueReconciliation = true
					containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_DEAD)
					break
				}

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

		ReconcileLoop(containerWatcher)

		break
	case status.STATUS_READY:
		containerObj.GetStatus().LastReadiness = true
		containerObj.GetStatus().LastReadinessTimestamp = time.Now()

		containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_RUNNING)
		ReconcileLoop(containerWatcher)

		break
	case status.STATUS_RUNNING:
		dockerState, err := containerObj.Get()

		if err != nil {
			ReconcileLoop(containerWatcher)
			break
		}

		switch dockerState.State {
		case "running":
			// shhhhh go to sleep
			containerWatcher.Logger.Debug("container is running")
			containerObj.GetStatus().Reconciling = false
			return
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
		case "stopped":
			containerWatcher.Logger.Error(fmt.Sprintf("%s container is stopped waiting for dead to restart", containerObj.GetGeneratedName()))
			break
		}

		ReconcileLoop(containerWatcher)
		break
	case status.STATUS_DEAD:
		dockerState, err := containerObj.Get()

		if err != nil {
			containerWatcher.Logger.Info(err.Error())
			break
		}

		switch dockerState.State {
		case "exited":
			containerWatcher.Logger.Info("container is dead")
			shared.Registry.BackOffTracking(containerObj.GetGroup(), containerObj.GetGeneratedName())

			if shared.Registry.BackOffTracker[containerObj.GetGroup()][containerObj.GetGeneratedName()] > 5 {
				containerWatcher.Logger.Error(fmt.Sprintf("%s container is backoff restarting", containerObj.GetGeneratedName()))

				shared.Registry.BackOffReset(containerObj.GetGroup(), containerObj.GetGeneratedName())

				containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_BACKOFF)
			} else {
				containerWatcher.Logger.Info("deleting dead container")

				err = containerObj.Delete()
				if err != nil {
					containerWatcher.Logger.Error(err.Error())
				}

				containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_PREPARE)
			}
			break
		case "created":
			containerWatcher.Logger.Info("container couldn't be created")
			shared.Registry.BackOffTracking(containerObj.GetGroup(), containerObj.GetGeneratedName())

			containerObj.Delete()
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_PREPARE)
			break
		default:
			containerWatcher.Logger.Info("container not dead retry again", zap.String("current-state", dockerState.State))
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_KILL)
		}

		ReconcileLoop(containerWatcher)
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
			go dependency.Ready(shared.Registry, containerObj.GetGroup(), containerObj.GetGeneratedName(), containerObj.GetDefinition().Spec.Container.Dependencies, containerWatcher.DependencyChan)

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
			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_DEAD)
		}

		containerWatcher.Container.GetStatus().Reconciling = false
		break

	case status.STATUS_KILL:
		containerWatcher.Logger.Info("attempt to shutdown gracefully")

		go func() {
			err := containerObj.Stop(static.SIGTERM)
			if err != nil {
				containerWatcher.Logger.Error(err.Error())
			}
		}()

		err := Wait(10*time.Second, func() error {
			c, err := containerObj.Get()

			if err != nil {
				return nil
			}

			if c != nil && c.State == "exited" {
				containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_DEAD)
			} else {
				return errors.New("container is not yet in exited state")
			}

			return nil
		})

		if err != nil {
			containerWatcher.Logger.Info("graceful shutdown failed - forcing kill")

			go func() {
				err = containerObj.Kill(static.SIGKILL)
				if err != nil {
					containerWatcher.Logger.Error(err.Error())
				}
			}()
			err = Wait(10*time.Second, func() error {
				var c *types.Container
				c, err = containerObj.Get()

				if err != nil {
					return err
				}

				if c != nil && c.State == "exited" {
					containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_DEAD)
				} else {
					return errors.New("container is not yet in exited state")
				}

				return nil
			})
		}

		if err != nil {
			containerWatcher.Logger.Error("failed to stop container - will try again")
		}

		ReconcileLoop(containerWatcher)
		break
	case status.STATUS_PENDING_DELETE:
		c, err := containerObj.Get()

		containerWatcher.Logger.Info("container is pending delete")

		if err != nil {
			shared.Registry.Remove(containerObj.GetGroup(), containerObj.GetGeneratedName())
			containerWatcher.Cancel()
		} else {
			if c.State == "running" {
				containerWatcher.Logger.Info("starting graceful termination, timeout 30s")

				err = containerObj.Stop(static.SIGTERM)

				if err != nil {
					containerWatcher.Logger.Error(err.Error())
				}
			}

			err = containerObj.Delete()

			if err != nil {
				containerWatcher.Logger.Info("failed to delete container from docker daemon")

			} else {
				containerWatcher.Logger.Info("container is deleted")

				shared.Registry.Remove(containerObj.GetGroup(), containerObj.GetGeneratedName())
				containerWatcher.Cancel()
			}
		}
		break
	}
}

func ReconcileLoop(containerWatcher *watcher.Container) {
	containerWatcher.Container.GetStatus().Reconciling = false
	containerWatcher.ContainerQueue <- containerWatcher.Container
}

func GetState(containerWatcher *watcher.Container) string {
	c, err := containerWatcher.Container.Get()

	if err != nil {
		return ""
	}

	if c != nil {
		return c.State
	}

	return ""
}

func Wait(timeout time.Duration, f func() error) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	backOff := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)

	return backoff.Retry(f, backOff)
}
