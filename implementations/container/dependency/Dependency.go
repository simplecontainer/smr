package dependency

import (
	"context"
	"errors"
	"github.com/cenkalti/backoff/v4"
	"github.com/simplecontainer/smr/implementations/container/shared"
	"github.com/simplecontainer/smr/implementations/container/status"
	"github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
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
		Timeout: depend.Timeout,
		Ctx:     ctx,
		Cancel:  cancel,
	}
}

func Ready(shared *shared.Shared, group string, name string, dependsOn []v1.DependsOn) (bool, error) {
	for _, depend := range dependsOn {
		dependency := NewDependencyFromDefinition(depend)
		dependency.Function = func() error {
			return SolveDepends(shared, group, name, dependency)
		}

		backOff := backoff.WithContext(backoff.NewExponentialBackOff(), dependency.Ctx)

		err := backoff.Retry(dependency.Function, backOff)
		if err != nil {
			container := shared.Registry.Find(group, name)

			if container != nil {
				container.Status.TransitionState(name, status.STATUS_DEPENDS_FAILED)
			}

			return false, nil
		}

		container := shared.Registry.Find(group, name)

		if container != nil {
			container.Status.TransitionState(name, status.STATUS_DEPENDS_SOLVED)
		}

		return true, nil
	}

	container := shared.Registry.Find(group, name)

	if container != nil {
		container.Status.TransitionState(name, status.STATUS_DEPENDS_SOLVED)
	}

	return true, nil
}

func SolveDepends(shared *shared.Shared, myGroup string, myName string, depend *Dependency) error {
	format := f.NewFromString(depend.Name)

	myContainer := shared.Registry.Find(myGroup, myName)

	if myContainer == nil {
		depend.Cancel()
	}

	otherGroup := format.Kind
	otherName := format.Group

	container := shared.Registry.Find(otherGroup, otherName)

	if container == nil {
		return errors.New("container not found")
	} else {
		if container.Status.LastReadiness {
			return nil
		} else {
			return errors.New("container not ready")
		}
	}
}
