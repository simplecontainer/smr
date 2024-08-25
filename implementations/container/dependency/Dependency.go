package dependency

import (
	"context"
	"errors"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/simplecontainer/smr/implementations/container/registry"
	"github.com/simplecontainer/smr/pkg/definitions/v1"
	"time"
)

func NewDependencyFromDefinition(depend v1.DependsOn) *Dependency {
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

func Ready(registry *registry.Registry, group string, name string, dependsOn []v1.DependsOn, channel chan *State) (bool, error) {
	for _, depend := range dependsOn {
		dependency := NewDependencyFromDefinition(depend)
		dependency.Function = func() error {
			return SolveDepends(registry, group, name, dependency, channel)
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

	if myContainer == nil { // || myContainer.Status.IfStateIs(status.STATUS_DEPENDS_CHECKING) {
		depend.Cancel()
		return errors.New("container not found")
	}

	otherGroup := depend.Group
	otherName := depend.Name

	if otherName == "*" {
		for _, container := range registry.Containers[otherGroup] {
			if container == nil {
				channel <- &State{
					State: CHECKING,
					Error: errors.New(fmt.Sprintf("container not found %s", container.Static.GeneratedName)),
				}

				return errors.New(fmt.Sprintf("container not found %s", container.Static.GeneratedName))
			} else {
				if !container.Status.LastReadiness {
					channel <- &State{
						State: CHECKING,
						Error: errors.New(fmt.Sprintf("container not ready %s", container.Static.GeneratedName)),
					}

					return errors.New(fmt.Sprintf("container not ready %s", container.Static.GeneratedName))
				}

				// Otherwise no-op continue to check next container and if every container is ready return nil
			}
		}

		return nil
	} else {
		container := registry.Find(otherGroup, otherName)

		if container == nil {
			channel <- &State{
				State: CHECKING,
				Error: errors.New(fmt.Sprintf("container not found %s.%s", otherGroup, otherName)),
			}

			return errors.New(fmt.Sprintf("container not found %s.%s", otherGroup, otherName))
		} else {
			if container.Status.LastReadiness {
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
