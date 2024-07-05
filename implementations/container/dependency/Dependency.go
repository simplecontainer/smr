package dependency

import (
	"context"
	"errors"
	"github.com/simplecontainer/smr/implementations/container/shared"
	"github.com/simplecontainer/smr/implementations/container/status"
	"github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/utils"
	"go.uber.org/zap"
	"time"
)

func NewDependencyFromDefinition(depend v1.DependsOn) *Dependency {
	if depend.Timeout == "" {
		depend.Timeout = "30s"
	}

	timeout, err := time.ParseDuration(depend.Timeout)

	var ctx context.Context
	if err == nil {
		ctx, _ = context.WithTimeout(context.Background(), timeout)
	} else {
		return nil
	}

	return &Dependency{
		Name:    depend.Name,
		Timeout: depend.Timeout,
		Ctx:     ctx,
	}
}

func Ready(shared *shared.Shared, group string, name string, dependsOn []v1.DependsOn) (bool, error) {
	if len(dependsOn) > 0 {
		var allDependenciesSolved = true
		logger.Log.Info("trying to solve dependencies", zap.String("group", group), zap.String("name", name))

		c := make(chan State)
		for _, depend := range dependsOn {
			dependency := NewDependencyFromDefinition(depend)
			go SolveDepends(shared, dependency, c)
		}

		for len(dependsOn) > 0 {
			select {
			case d := <-c:
				if d.Missing {
					allDependenciesSolved = false

					for i, v := range dependsOn {
						if v.Name == d.Depend.Name {
							dependsOn = append(dependsOn[:i], dependsOn[i+1:]...)
						}
					}
				}

				if d.Success {
					logger.Log.Info("dependency solved", zap.String("group", group), zap.String("name", name))

					for i, v := range dependsOn {
						if v.Name == d.Depend.Name {
							dependsOn = append(dependsOn[:i], dependsOn[i+1:]...)
						}
					}
				} else {
					deadline, _ := d.Depend.Ctx.Deadline()

					if deadline.After(time.Now()) {
						time.Sleep(5 * time.Second)
						go SolveDepends(shared, d.Depend, c)
					} else {
						logger.Log.Info("deadline exceeded", zap.String("group", group), zap.String("name", name))
						allDependenciesSolved = false

						for i, v := range dependsOn {
							if v.Name == d.Depend.Name {
								dependsOn = append(dependsOn[:i], dependsOn[i+1:]...)
							}
						}
					}
				}
			}
		}

		if !allDependenciesSolved {
			shared.Registry.Containers[group][name].Status.TransitionState(name, status.STATUS_DEPENDS_FAILED)
			return false, errors.New("didn't solve all dependencies")
		} else {
			shared.Registry.Containers[group][name].Status.TransitionState(name, status.STATUS_DEPENDS_SOLVED)
			logger.Log.Info("all dependencies solved", zap.String("group", group), zap.String("name", name))
			return true, nil
		}
	}

	logger.Log.Info("no dependencies defined", zap.String("group", group), zap.String("name", name))
	shared.Registry.Containers[group][name].Status.TransitionState(name, status.STATUS_DEPENDS_SOLVED)

	return true, nil
}

func SolveDepends(shared *shared.Shared, depend *Dependency, c chan State) {
	if depend.Timeout == "" {
		depend.Timeout = "30s"
	}

	timeout, err := time.ParseDuration(depend.Timeout)

	if err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		ch := make(chan State)
		defer close(ch)

		logger.Log.Info("trying to solve dependency", zap.String("name", depend.Name))

		go Depends(shared, depend, ch)

		for {
			select {
			case d := <-ch:
				c <- d
			case <-ctx.Done():
				c <- State{
					Success: false,
					Missing: false,
					Timeout: true,
					Depend:  depend,
				}
			}
		}
	} else {
		c <- State{
			Success: false,
			Missing: false,
			Timeout: false,
			Error:   err,
			Depend:  depend,
		}
	}
}

func Depends(shared *shared.Shared, depend *Dependency, ch chan State) {
	group, id := utils.ExtractGroupAndId(depend.Name)

	logger.Log.Info("trying to check if depends solved", zap.String("group", group), zap.String("name", id))

	if shared.Registry.Containers[group] != nil {
		if id == "*" {
			for _, container := range shared.Registry.Containers[group] {
				if !container.Status.IfStateIs(status.STATUS_READY) {
					ch <- State{
						Success: false,
						Depend:  depend,
					}

					return
				}
			}

			ch <- State{
				Success: true,
				Missing: false,
				Depend:  depend,
			}

			return
		} else {
			if shared.Registry.Containers[group][id] != nil {
				ch <- State{
					Success: true,
					Missing: false,
					Depend:  depend,
				}
			} else {
				ch <- State{
					Success: false,
					Missing: true,
					Depend:  depend,
				}
			}
		}
	} else {
		ch <- State{
			Success: false,
			Missing: true,
			Depend:  depend,
		}
	}

}
