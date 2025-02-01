package dependency

import (
	"context"
	"errors"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/kinds/container/registry"
	"github.com/simplecontainer/smr/pkg/logger"
	"time"
)

func NewDependencyFromDefinition(depend v1.ContainerDependsOn) *Dependency {
	if depend.Timeout == "" {
		depend.Timeout = "30s"
	}

	timeout, err := time.ParseDuration(depend.Timeout)

	if err != nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	return &Dependency{
		Prefix:  depend.Prefix,
		Name:    depend.Name,
		Group:   depend.Group,
		Timeout: depend.Timeout,
		Ctx:     ctx,
		Cancel:  cancel,
	}
}

func Ready(registry *registry.Registry, prefix string, group string, name string, dependsOn []v1.ContainerDependsOn, channel chan *State) (bool, error) {
	for _, depend := range dependsOn {
		dependency := NewDependencyFromDefinition(depend)
		dependency.Function = func() error {
			err := SolveDepends(registry, prefix, group, name, dependency, channel)

			if err != nil {
				logger.Log.Info(err.Error())
			}

			return err
		}

		backOff := backoff.WithContext(backoff.NewExponentialBackOff(), dependency.Ctx)

		err := backoff.Retry(dependency.Function, backOff)
		if err != nil {
			dependency.Cancel()

			channel <- &State{
				State: FAILED,
				Error: err,
			}

			return false, err
		}
	}

	channel <- &State{
		State: SUCCESS,
		Error: nil,
	}

	return true, nil
}

func SolveDepends(registry *registry.Registry, myPrefix string, myGroup string, myName string, depend *Dependency, channel chan *State) error {
	myContainer := registry.Find(myPrefix, myGroup, myName)

	if myContainer == nil {
		depend.Cancel()
		return errors.New("container not found")
	}

	otherPrefix := depend.Prefix
	otherGroup := depend.Group
	otherName := depend.Name

	if otherName == "*" {
		containers := registry.FindGroup(otherPrefix, otherGroup)

		if len(containers) == 0 {
			return errors.New("waiting for atleast one container from group to show up")
		} else {
			flagFail := false

			for _, container := range containers {
				if !container.GetStatus().LastReadiness {
					channel <- &State{
						State: CHECKING,
						Error: errors.New(fmt.Sprintf("container not ready %s", container.GetGeneratedName())),
					}

					flagFail = true
					break
				}
			}

			if flagFail {
				return errors.New("dependency not ready")
			} else {
				return nil
			}
		}
	} else {
		container := registry.Find(otherPrefix, otherGroup, otherName)

		if container == nil {
			channel <- &State{
				State: CHECKING,
				Error: errors.New(fmt.Sprintf("container not found %s.%s", otherGroup, otherName)),
			}

			return errors.New(fmt.Sprintf("container not found %s.%s", otherGroup, otherName))
		} else {
			if container.GetStatus().LastReadiness {
				return nil
			} else {
				channel <- &State{
					State: CHECKING,
					Error: errors.New("container not ready"),
				}

				return errors.New("container not ready")
			}
		}
	}
}
