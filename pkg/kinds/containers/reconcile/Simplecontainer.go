package reconcile

import (
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/dependency"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/readiness/solver"
	"github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/kinds/containers/status"
	"github.com/simplecontainer/smr/pkg/kinds/containers/watcher"
	"github.com/simplecontainer/smr/pkg/static"
	"reflect"
	"time"
)

func Reconcile(shared *shared.Shared, containerWatcher *watcher.Container, existing platforms.IContainer, engine string, engineError string) (string, bool) {
	containerObj := containerWatcher.Container

	switch containerObj.GetStatus().State.State {
	case status.STATUS_CLEAN:
		switch engine {
		case "exited":
			err := containerObj.Delete()
			if err != nil {
				containerWatcher.Logger.Error(err.Error())
				return status.STATUS_DAEMON_FAILURE, true
			}

			if containerObj.GetStatus().State.PreviousState != status.STATUS_CLEAN {
				return containerObj.GetStatus().State.PreviousState, true
			} else {
				return status.STATUS_DEAD, true
			}
		default:
			err := containerObj.Stop(static.SIGTERM)
			if err != nil {
				err = containerObj.Kill(static.SIGKILL)

				if err != nil {
					containerWatcher.Logger.Error(err.Error())
					return status.STATUS_DAEMON_FAILURE, true
				}
			}

			return "", false
		}
	case status.STATUS_TRANSFERING:
		if existing != nil && existing.IsGhost() {
			containerWatcher.Logger.Info("container is not dead on another node - wait")
			return status.STATUS_TRANSFERING, false
		} else {
			containerWatcher.Logger.Info("container transefered on this node")
			return status.STATUS_CREATED, true
		}
	case status.STATUS_CREATED:
		if engine != "" {
			return status.STATUS_CLEAN, true
		} else {
			return status.STATUS_PREPARE, true
		}
	case status.STATUS_PREPARE:
		err := containerObj.PreRun(shared.Manager.Config, shared.Client, containerWatcher.User)

		if err != nil {
			containerWatcher.Logger.Error(err.Error())
			return status.STATUS_PENDING, true
		} else {
			go func() {
				_, err = dependency.Ready(containerWatcher.Ctx, shared.Registry, containerObj.GetGroup(), containerObj.GetGeneratedName(), containerObj.GetDefinition().(*v1.ContainersDefinition).Spec.Dependencies, containerWatcher.DependencyChan)

				if err != nil {
					containerWatcher.Logger.Error(err.Error())
				}
			}()

			containerWatcher.Logger.Info("container prepared")
			return status.STATUS_DEPENDS_CHECKING, true
		}
	case status.STATUS_PENDING:
		err := containerObj.PreRun(shared.Manager.Config, shared.Client, containerWatcher.User)

		if err != nil {
			containerWatcher.Logger.Error(err.Error())
			return status.STATUS_PENDING, false
		} else {
			go func() {
				_, err = dependency.Ready(containerWatcher.Ctx, shared.Registry, containerObj.GetGroup(), containerObj.GetGeneratedName(), containerObj.GetDefinition().(*v1.ContainersDefinition).Spec.Dependencies, containerWatcher.DependencyChan)
				if err != nil {
					containerWatcher.Logger.Error(err.Error())
				}
			}()

			containerWatcher.Logger.Info("container prepared")
			return status.STATUS_DEPENDS_CHECKING, true
		}
	case status.STATUS_DEPENDS_CHECKING:
		for {
			select {
			case dependencyResult := <-containerWatcher.DependencyChan:
				if dependencyResult == nil {
					return status.STATUS_DEPENDS_FAILED, true
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
					return status.STATUS_DEPENDS_SOLVED, true
				case dependency.FAILED:
					containerWatcher.Logger.Info("dependency check failed")
					return status.STATUS_DEPENDS_FAILED, true
				}
			}
		}
	case status.STATUS_DEPENDS_SOLVED:
		containerObj.GetStatus().LastDependsSolved = true
		containerObj.GetStatus().LastDependsSolvedTimestamp = time.Now()

		if !reflect.ValueOf(containerObj.GetInitDefinition()).IsZero() {
			return status.STATUS_INIT, true
		} else {
			return status.STATUS_START, true
		}
	case status.STATUS_INIT:
		err := containerObj.InitContainer(containerObj.GetInitDefinition(), shared.Manager.Config, shared.Client, containerWatcher.User)

		if err != nil {
			return status.STATUS_INIT_FAILED, true
		} else {
			return status.STATUS_START, true
		}
	case status.STATUS_INIT_FAILED:
		containerWatcher.Logger.Info("init container exited with error - reconciler going to sleep")
		return status.STATUS_INIT_FAILED, false
	case status.STATUS_DEPENDS_FAILED:
		containerWatcher.Logger.Info("container depends timeout or failed - retry again")
		return status.STATUS_PREPARE, true
	case status.STATUS_START:
		containerWatcher.Logger.Info("container attempt to start")
		err := containerObj.Run()

		if err != nil {
			containerWatcher.Logger.Error(err.Error())
			return status.STATUS_DAEMON_FAILURE, true
		} else {
			containerWatcher.Logger.Info("container started")

			go func() {
				_, err = solver.Ready(containerWatcher.Ctx, shared.Client, containerObj, containerWatcher.User, containerWatcher.ReadinessChan, containerWatcher.Logger)

				if err != nil {
					containerWatcher.Logger.Error(err.Error())
				}
			}()

			return status.STATUS_READINESS_CHECKING, true
		}
	case status.STATUS_READINESS_CHECKING:
		for {
			select {
			case readinessResult := <-containerWatcher.ReadinessChan:
				state := GetState(containerWatcher)

				if state.State == "running" {
					switch readinessResult.State {
					case dependency.CHECKING:
						containerWatcher.Logger.Info("checking readiness")
						break
					case dependency.SUCCESS:
						containerWatcher.Logger.Info("readiness check success")

						// Add dns only when container is ready so no one can contact it earlier
						err := containerObj.PostRun(shared.Manager.Config, shared.Manager.DnsCache)

						if err != nil {
							containerWatcher.Logger.Info("container failed to update dns - proceed with restart")
							containerWatcher.Logger.Error(err.Error())

							return status.STATUS_KILL, true
						} else {
							return status.STATUS_READY, true
						}
					case dependency.FAILED:
						containerWatcher.Logger.Info("readiness check failed")
						break
					}
				} else {
					return status.STATUS_READINESS_FAILED, true
				}
			}
		}
	case status.STATUS_READY:
		containerObj.GetStatus().LastReadiness = true
		containerObj.GetStatus().LastReadinessTimestamp = time.Now()

		return status.STATUS_RUNNING, true

	case status.STATUS_READINESS_FAILED:
		containerWatcher.Logger.Info("container readiness failed")
		return status.STATUS_KILL, true

	case status.STATUS_RUNNING:
		containerWatcher.Logger.Info("container is running - reconciler going to sleep")
		return status.STATUS_RUNNING, false

	case status.STATUS_KILL:
		err := containerObj.Stop(static.SIGTERM)

		if err != nil {
			err = containerObj.Stop(static.SIGKILL)

			if err != nil {
				containerWatcher.Logger.Error(err.Error())
			}
		}

		return status.STATUS_DEAD, false
	case status.STATUS_DEAD:
		if err := shared.Registry.BackOff(containerObj.GetGroup(), containerObj.GetGeneratedName()); err != nil {
			return status.STATUS_BACKOFF, true
		} else {
			containerWatcher.Logger.Info("deleting dead container")

			err := containerObj.Delete()
			if err != nil {
				containerWatcher.Logger.Error(err.Error())
			}

			return status.STATUS_PREPARE, true
		}

	case status.STATUS_PENDING_DELETE:
		containerWatcher.Container.GetStatus().PendingDelete = true

		if containerObj.GetStatus().State.PreviousState == status.STATUS_CLEAN || engine == "exited" {
			state, err := containerObj.GetState()
			if err == nil && state.State == "exited" {
				if err := containerObj.Delete(); err != nil {
					containerWatcher.Logger.Error(err.Error())
					return status.STATUS_DAEMON_FAILURE, true
				}
			}

			return status.STATUS_PENDING_DELETE, false
		}

		return status.STATUS_CLEAN, true
	case status.STATUS_DAEMON_FAILURE:
		containerWatcher.Logger.Info("container daemon engine failed - reconciler going to sleep")
		return status.STATUS_DAEMON_FAILURE, false

	case status.STATUS_BACKOFF:
		containerWatcher.Logger.Info("container is in backoff - reconciler going to sleep")
		return status.STATUS_BACKOFF, false

	default:
		return status.STATUS_CREATED, true
	}
}
