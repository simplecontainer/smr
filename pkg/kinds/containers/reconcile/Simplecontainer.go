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

	containerWatcher.AllowPlatformEvents = false

	switch containerObj.GetStatus().State.State {
	case status.TRANSFERING:
		if existing != nil && existing.IsGhost() {
			containerWatcher.Logger.Info("container is not dead on another node - wait")
			return status.TRANSFERING, false
		} else {
			containerWatcher.Logger.Info("container transefered on this node")
			return status.CREATED, true
		}
	case status.RESTART:
		containerWatcher.Logger.Info("container is restarted")
		return status.CLEAN, true
	case status.CREATED:
		containerWatcher.Logger.Info("container is created")
		return status.CLEAN, true
	case status.CLEAN:
		err := containerObj.Clean()

		if err != nil {
			containerWatcher.Logger.Error(err.Error())
			return status.DAEMON_FAILURE, true
		}

		return status.PREPARE, true
	case status.PREPARE:
		err := containerObj.PreRun(shared.Manager.Config, shared.Client, containerWatcher.User)

		if err != nil {
			containerWatcher.Logger.Error(err.Error())
			return status.PENDING, true
		} else {
			go func() {
				_, err = dependency.Ready(containerWatcher.Ctx, shared.Registry, containerObj.GetGroup(), containerObj.GetGeneratedName(), containerObj.GetDefinition().(*v1.ContainersDefinition).Spec.Dependencies, containerWatcher.DependencyChan)

				if err != nil {
					containerWatcher.Logger.Error(err.Error())
				}
			}()

			containerWatcher.Logger.Info("container prepared")
			return status.DEPENDS_CHECKING, true
		}
	case status.PENDING:
		err := containerObj.PreRun(shared.Manager.Config, shared.Client, containerWatcher.User)

		if err != nil {
			containerWatcher.Logger.Error(err.Error())
			return status.PENDING, false
		} else {
			go func() {
				_, err = dependency.Ready(containerWatcher.Ctx, shared.Registry, containerObj.GetGroup(), containerObj.GetGeneratedName(), containerObj.GetDefinition().(*v1.ContainersDefinition).Spec.Dependencies, containerWatcher.DependencyChan)
				if err != nil {
					containerWatcher.Logger.Error(err.Error())
				}
			}()

			containerWatcher.Logger.Info("container prepared")
			return status.DEPENDS_CHECKING, true
		}
	case status.DEPENDS_CHECKING:
		for {
			select {
			case dependencyResult := <-containerWatcher.DependencyChan:
				if dependencyResult == nil {
					return status.DEPENDS_FAILED, true
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
					return status.DEPENDS_SOLVED, true
				case dependency.FAILED:
					containerWatcher.Logger.Info("dependency check failed")
					return status.DEPENDS_FAILED, true
				}
			}
		}
	case status.DEPENDS_SOLVED:
		containerObj.GetStatus().LastDependsSolved = true
		containerObj.GetStatus().LastDependsSolvedTimestamp = time.Now()

		if !reflect.ValueOf(containerObj.GetInitDefinition()).IsZero() {
			return status.INIT, true
		} else {
			return status.START, true
		}
	case status.INIT:
		err := containerObj.InitContainer(containerObj.GetInitDefinition(), shared.Manager.Config, shared.Client, containerWatcher.User)

		if err != nil {
			return status.INIT_FAILED, true
		} else {
			return status.START, true
		}
	case status.INIT_FAILED:
		containerWatcher.Logger.Info("init container exited with error - reconciler going to sleep")
		return status.INIT_FAILED, false
	case status.DEPENDS_FAILED:
		containerWatcher.Logger.Info("container depends timeout or failed - retry again")
		return status.PREPARE, true
	case status.START:
		containerWatcher.Logger.Info("container attempt to start")
		err := containerObj.Run()

		if err != nil {
			containerWatcher.Logger.Error(err.Error())
			return status.DAEMON_FAILURE, true
		} else {
			containerWatcher.Logger.Info("container started")

			go func() {
				_, err = solver.Ready(containerWatcher.Ctx, shared.Client, containerObj, containerWatcher.User, containerWatcher.ReadinessChan, containerWatcher.Logger)

				if err != nil {
					containerWatcher.Logger.Error(err.Error())
				}
			}()

			return status.READINESS_CHECKING, true
		}
	case status.READINESS_CHECKING:
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

							return status.KILL, true
						} else {
							containerWatcher.Logger.Info("container updated dns")
							return status.READY, true
						}
					case dependency.FAILED:
						containerWatcher.Logger.Info("readiness check failed")
						break
					}
				} else {
					return status.READINESS_FAILED, true
				}
			}
		}
	case status.READY:
		containerObj.GetStatus().LastReadiness = true
		containerObj.GetStatus().LastReadinessTimestamp = time.Now()

		return status.RUNNING, true
	case status.READINESS_FAILED:
		containerWatcher.Logger.Info("container readiness failed")
		return status.KILL, true

	case status.RUNNING:
		containerWatcher.AllowPlatformEvents = true
		containerWatcher.Logger.Info("container is running - reconciler going to sleep")
		return status.RUNNING, false

	case status.KILL:
		err := containerObj.Stop(static.SIGTERM)

		if err != nil {
			err = containerObj.Stop(static.SIGKILL)

			if err != nil {
				containerWatcher.Logger.Error(err.Error())
			}
		}

		return status.DEAD, false
	case status.DEAD:
		if err := shared.Registry.BackOff(containerObj.GetGroup(), containerObj.GetGeneratedName()); err != nil {
			return status.BACKOFF, true
		} else {
			containerWatcher.Logger.Info("deleting dead container")

			err = containerObj.Clean()

			if err != nil {
				containerWatcher.Logger.Error(err.Error())
				return status.DAEMON_FAILURE, true
			}

			return status.PREPARE, true
		}

	case status.DELETE:
		return "", false
	case status.DAEMON_FAILURE:
		containerWatcher.Logger.Info("container daemon engine failed - reconciler going to sleep")
		return status.DAEMON_FAILURE, false

	case status.BACKOFF:
		containerWatcher.Logger.Info("container is in backoff - reconciler going to sleep")
		return status.BACKOFF, false

	default:
		return status.CREATED, true
	}
}
