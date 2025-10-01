package dependency

import (
	"context"
	"errors"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"go.uber.org/zap"
	"time"
)

var (
	ERROR_CONTEXT_CANCELED        = errors.New("context canceled")
	ERROR_CONTEXT_TIMEOUT         = errors.New("context time out")
	ERROR_CONTAINER_NOT_FOUND     = errors.New("container not found")
	ERROR_WAITING_FOR_ATLEAST_ONE = errors.New("waiting for atleast one container from group to show up")
	ERROR_DEPENDENCY_NOT_READY    = errors.New("dependency not ready")
	ERROR_CONTAINER_NOT_READY     = errors.New("container not ready")
)

func NewDependencyFromDefinition(depend v1.ContainersDependsOn) *Dependency {
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

func Ready(ctx context.Context, registry platforms.Registry, group string, name string, dependsOn []v1.ContainersDependsOn, channel chan *State, logger *zap.Logger) (bool, error) {
	for _, depend := range dependsOn {
		dependency := NewDependencyFromDefinition(depend)
		dependency.Function = func() error {
			if ctx.Err() != nil {
				return backoff.Permanent(ctx.Err())
			}

			if dependency.Ctx.Err() != nil {
				return backoff.Permanent(dependency.Ctx.Err())
			}

			err := SolveDepends(registry, depend.Prefix, group, name, dependency, channel, logger)

			if err != nil {
				if errors.Is(err, ERROR_CONTAINER_NOT_FOUND) {
					return backoff.Permanent(err)
				} else {
					channel <- &State{
						State: CHECKING,
						Error: err,
					}

					return err
				}
			}

			return nil
		}

		backOff := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
		err := backoff.Retry(dependency.Function, backOff)

		if ctx.Err() != nil {
			return false, ctx.Err()
		}

		if dependency.Ctx.Err() != nil {
			channel <- &State{
				State: CANCELED,
				Error: dependency.Ctx.Err(),
			}
			return false, dependency.Ctx.Err()
		}

		if err != nil {
			channel <- &State{
				State: FAILED,
				Error: err,
			}

			return false, err
		}
	}

	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	channel <- &State{
		State: SUCCESS,
		Error: nil,
	}

	return true, nil
}

func SolveDepends(registry platforms.Registry, myPrefix string, myGroup string, myName string, depend *Dependency, channel chan *State, logger *zap.Logger) error {
	myContainer := registry.Find(myPrefix, myGroup, myName)
	if myContainer == nil {
		return ERROR_CONTAINER_NOT_FOUND
	}

	otherPrefix := depend.Prefix
	otherGroup := depend.Group
	otherName := depend.Name

	if otherName == "*" {
		containers := registry.FindGroup(otherPrefix, otherGroup)

		if len(containers) == 0 {
			return ERROR_WAITING_FOR_ATLEAST_ONE
		}

		flagFail := false
		for _, container := range containers {
			if !container.GetStatus().LastReadiness {
				channel <- &State{
					State: CHECKING,
					Error: fmt.Errorf("container not ready %s", container.GetGeneratedName()),
				}
				flagFail = true
				break
			}
		}

		if flagFail {
			return ERROR_DEPENDENCY_NOT_READY
		}
		return nil

	} else {
		container := registry.Find(otherPrefix, otherGroup, otherName)
		if container == nil {
			logger.Error("container not found", zap.String("group", otherGroup), zap.String("name", otherName))
			channel <- &State{
				State: CHECKING,
				Error: nil,
			}
			return ERROR_CONTAINER_NOT_FOUND
		}

		if container.GetStatus().LastReadiness {
			return nil
		}

		channel <- &State{
			State: CHECKING,
			Error: ERROR_CONTAINER_NOT_READY,
		}
		return ERROR_CONTAINER_NOT_READY
	}
}
