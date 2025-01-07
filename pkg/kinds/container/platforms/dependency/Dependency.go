package dependency

import (
	"context"
	"errors"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/kinds/container/registry"
	"github.com/simplecontainer/smr/pkg/logger"
	"sort"
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
		Name:    depend.Name,
		Group:   depend.Group,
		Timeout: depend.Timeout,
		Ctx:     ctx,
		Cancel:  cancel,
	}
}

func Ready(registry *registry.Registry, group string, name string, dependsOn []v1.ContainerDependsOn, channel chan *State) (bool, error) {
	for _, depend := range dependsOn {
		dependency := NewDependencyFromDefinition(depend)
		dependency.Function = func() error {
			err := SolveDepends(registry, group, name, dependency, channel)

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

func SolveDepends(registry *registry.Registry, myGroup string, myName string, depend *Dependency, channel chan *State) error {
	myContainer := registry.Find(myGroup, myName)

	if myContainer == nil {
		depend.Cancel()
		return errors.New("container not found")
	}

	otherGroup := depend.Group
	otherName := depend.Name

	if otherName == "*" {
		containers := registry.FindGroup(otherGroup)
		keys := make([]string, 0)

		for k, _ := range containers {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		if len(containers) == 0 {
			return errors.New("waiting for atleast one container from group to show up")
		} else {
			flagFail := false

			for _, containerName := range keys {
				if containers[containerName] == nil {
					channel <- &State{
						State: CHECKING,
						Error: errors.New(fmt.Sprintf("container not found %s", containers[containerName].GetGeneratedName())),
					}

					flagFail = true
				} else {
					if !containers[containerName].GetStatus().LastReadiness {
						channel <- &State{
							State: CHECKING,
							Error: errors.New(fmt.Sprintf("container not ready %s", containers[containerName].GetGeneratedName())),
						}

						flagFail = true
					}
				}
			}

			if flagFail {
				return errors.New("dependency not ready")
			} else {
				return nil
			}
		}
	} else {
		container := registry.Find(otherGroup, otherName)

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
