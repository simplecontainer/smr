package reconcile

import (
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/image"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/readiness"
	"reflect"
	"time"

	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/dependency"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/readiness/solver"
	"github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/kinds/containers/status"
	"github.com/simplecontainer/smr/pkg/kinds/containers/watcher"
	"github.com/simplecontainer/smr/pkg/static"
)

type StateHandlerFunc func(shared *shared.Shared, cw *watcher.Container, existing platforms.IContainer) (string, bool)

var stateHandlers = map[string]StateHandlerFunc{
	status.TRANSFERING:        handleTransferring,
	status.RESTART:            handleRestart,
	status.CREATED:            handleCreated,
	status.CLEAN:              handleClean,
	status.PREPARE:            handlePrepare,
	status.PENDING:            handlePending,
	status.DEPENDS_CHECKING:   handleDependsChecking,
	status.DEPENDS_SOLVED:     handleDependsSolved,
	status.INIT:               handleInit,
	status.INIT_FAILED:        handleInitFailed,
	status.DEPENDS_FAILED:     handleDependsFailed,
	status.START:              handleStart,
	status.READINESS_CHECKING: handleReadinessChecking,
	status.READY:              handleReady,
	status.READINESS_FAILED:   handleReadinessFailed,
	status.RUNNING:            handleRunning,
	status.KILL:               handleKill,
	status.DEAD:               handleDead,
	status.DELETE:             handleDelete,
	status.DAEMON_FAILURE:     handleDaemonFailure,
	status.BACKOFF:            handleBackoff,
}

func Reconcile(shared *shared.Shared, cw *watcher.Container, existing platforms.IContainer, engine string, engineError string) (string, bool) {
	state := cw.Container.GetStatus().State.State

	if handler, ok := stateHandlers[state]; ok {
		return handler(shared, cw, existing)
	}

	return status.CREATED, true
}

func handleTransferring(shared *shared.Shared, cw *watcher.Container, existing platforms.IContainer) (string, bool) {
	if existing != nil && existing.IsGhost() {
		cw.Logger.Info("container is not dead on another node - wait")
		return status.TRANSFERING, false
	}
	cw.Logger.Info("container transferred on this node")
	return status.CREATED, true
}

func handleRestart(shared *shared.Shared, cw *watcher.Container, existing platforms.IContainer) (string, bool) {
	cw.AllowPlatformEvents = false

	cw.Logger.Info("container is restarted")
	return status.CLEAN, true
}

func handleCreated(shared *shared.Shared, cw *watcher.Container, existing platforms.IContainer) (string, bool) {
	cw.Logger.Info("container is created")
	return status.CLEAN, true
}

func handleClean(shared *shared.Shared, cw *watcher.Container, existing platforms.IContainer) (string, bool) {
	cw.Logger.Info("container is cleaning old container")

	cw.AllowPlatformEvents = false
	if err := cw.Container.Clean(); err != nil {
		cw.Logger.Error(err.Error())
		return status.DAEMON_FAILURE, true
	}
	cw.AllowPlatformEvents = true

	return status.PREPARE, true
}

func handlePrepare(shared *shared.Shared, cw *watcher.Container, existing platforms.IContainer) (string, bool) {
	if err := cw.Container.PreRun(shared.Manager.Config, shared.Client, cw.User); err != nil {
		cw.Logger.Error(err.Error())
		return status.PENDING, true
	}

	go func() {
		_, err := dependency.Ready(cw.ReconcileCtx, shared.Registry, cw.Container.GetGroup(), cw.Container.GetGeneratedName(),
			cw.Container.GetDefinition().(*v1.ContainersDefinition).Spec.Dependencies, cw.DependencyChan)
		if err != nil {
			cw.Logger.Error(err.Error())
		}
	}()

	cw.Logger.Info("container prepared")
	return status.DEPENDS_CHECKING, true
}

func handleDependsChecking(shared *shared.Shared, cw *watcher.Container, existing platforms.IContainer) (string, bool) {
	for {
		select {
		case dependencyResult := <-cw.DependencyChan:
			if dependencyResult == nil {
				return status.DEPENDS_FAILED, true
			}

			switch dependencyResult.State {
			case dependency.CHECKING:
				cw.Logger.Info("checking dependency")

				if dependencyResult.Error != nil {
					cw.Logger.Info(dependencyResult.Error.Error())
				}

				break
			case dependency.SUCCESS:
				cw.Logger.Info("dependency check success")
				return status.DEPENDS_SOLVED, true
			case dependency.FAILED:
				cw.Logger.Info("dependency check failed")
				return status.DEPENDS_FAILED, true
			case dependency.CANCELED:
				cw.Logger.Info("dependency check canceled")
				return "", false
			}
		}
	}
}

func handleDependsSolved(shared *shared.Shared, cw *watcher.Container, existing platforms.IContainer) (string, bool) {
	cw.Container.GetStatus().LastDependsSolved = true
	cw.Container.GetStatus().LastDependsSolvedTimestamp = time.Now()

	if !reflect.ValueOf(cw.Container.GetInitDefinition()).IsZero() {
		return status.INIT, true
	}

	cw.Container.GetImageState().SetStatus(image.StatusPulling)
	return status.START, true
}

func handleInit(shared *shared.Shared, cw *watcher.Container, existing platforms.IContainer) (string, bool) {
	if err := cw.Container.InitContainer(cw.Container.GetInitDefinition(), shared.Manager.Config, shared.Client, cw.User); err != nil {
		return status.INIT_FAILED, true
	}

	return status.START, true
}

func handleInitFailed(shared *shared.Shared, cw *watcher.Container, existing platforms.IContainer) (string, bool) {
	cw.Logger.Info("init container exited with error - reconciler going to sleep")
	return status.INIT_FAILED, false
}

func handleDependsFailed(shared *shared.Shared, cw *watcher.Container, existing platforms.IContainer) (string, bool) {
	cw.Logger.Info("container depends timeout or failed - retry again")
	return status.PREPARE, true
}

func handleStart(shared *shared.Shared, cw *watcher.Container, existing platforms.IContainer) (string, bool) {
	cw.Logger.Info("container attempt to start")

	if err := cw.Container.Run(); err != nil {
		cw.Logger.Error(err.Error())
		return status.DAEMON_FAILURE, true
	}

	cw.Logger.Info("container started")

	go func() {
		_, err := solver.Ready(cw.ReconcileCtx, shared.Client, cw.Container, cw.User, cw.ReadinessChan, cw.Logger)
		if err != nil {
			cw.Logger.Error(err.Error())
		}
	}()

	return status.READINESS_CHECKING, true
}

func handleReadinessChecking(shared *shared.Shared, cw *watcher.Container, existing platforms.IContainer) (string, bool) {
	for {
		select {
		case readinessResult := <-cw.ReadinessChan:
			state := GetState(cw)

			if state.State == "running" {
				switch readinessResult.State {
				case readiness.CHECKING:
					cw.Logger.Info("checking readiness")
					break
				case readiness.SUCCESS:
					cw.Logger.Info("readiness check success")

					// Add dns only when container is ready so no one can contact it earlier
					err := cw.Container.PostRun(shared.Manager.Config, shared.Manager.DnsCache)

					if err != nil {
						cw.Logger.Info("container failed to update dns - proceed with restart")
						cw.Logger.Error(err.Error())

						return status.KILL, true
					} else {
						cw.Logger.Info("container updated dns")
						return status.READY, true
					}
				case readiness.FAILED:
					cw.Logger.Info("readiness check failed")
					return status.READINESS_FAILED, true
				case readiness.CANCELED:
					cw.Logger.Info("readiness check canceled")
					return "", false
				}
			} else {
				return status.READINESS_FAILED, true
			}
		}
	}
}

func handleReady(shared *shared.Shared, cw *watcher.Container, existing platforms.IContainer) (string, bool) {
	cw.Container.GetStatus().LastReadiness = true
	cw.Container.GetStatus().LastReadinessTimestamp = time.Now()
	return status.RUNNING, true
}

func handleReadinessFailed(shared *shared.Shared, cw *watcher.Container, existing platforms.IContainer) (string, bool) {
	cw.Logger.Info("container readiness failed")
	cw.Container.GetStatus().LastReadiness = false
	cw.Container.GetStatus().LastReadinessTimestamp = time.Now()
	return status.KILL, true
}

func handleRunning(shared *shared.Shared, cw *watcher.Container, existing platforms.IContainer) (string, bool) {
	shared.Registry.BackOffReset(cw.Container.GetGroup(), cw.Container.GetGeneratedName())
	cw.Logger.Info("container is running, backoff is cleared - reconciler going to sleep")
	return status.RUNNING, false
}

func handleKill(shared *shared.Shared, cw *watcher.Container, existing platforms.IContainer) (string, bool) {
	if err := cw.Container.Stop(static.SIGTERM); err != nil {
		if err = cw.Container.Stop(static.SIGKILL); err != nil {
			cw.Logger.Error(err.Error())
		}
	}
	return status.DEAD, false
}

func handleDead(shared *shared.Shared, cw *watcher.Container, existing platforms.IContainer) (string, bool) {
	if err := shared.Registry.BackOff(cw.Container.GetGroup(), cw.Container.GetGeneratedName()); err != nil {
		return status.BACKOFF, true
	}

	cw.Logger.Info("deleting dead container")

	if err := cw.Container.Clean(); err != nil {
		cw.Logger.Error(err.Error())
		return status.DAEMON_FAILURE, true
	}

	return status.PREPARE, true
}

func handleDelete(shared *shared.Shared, cw *watcher.Container, existing platforms.IContainer) (string, bool) {
	cw.Done = true
	cw.Cancel()
	return "", false
}

func handleDaemonFailure(shared *shared.Shared, cw *watcher.Container, existing platforms.IContainer) (string, bool) {
	cw.Logger.Info("container daemon engine failed - reconciler going to sleep")
	return status.DAEMON_FAILURE, false
}

func handleBackoff(shared *shared.Shared, cw *watcher.Container, existing platforms.IContainer) (string, bool) {
	cw.Logger.Info("container is in backoff - reconciler going to sleep")
	return status.BACKOFF, false
}

func handlePending(shared *shared.Shared, cw *watcher.Container, existing platforms.IContainer) (string, bool) {
	cw.Logger.Info("container is waiting for external updates - reconciler going to sleep")
	return status.PENDING, false
}
