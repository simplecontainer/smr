package dependency

import (
	"context"
	"errors"
	"github.com/qdnqn/smr/pkg/definitions/v1"
	"github.com/qdnqn/smr/pkg/logger"
	"github.com/qdnqn/smr/pkg/manager"
	"github.com/qdnqn/smr/pkg/status"
	"github.com/qdnqn/smr/pkg/utils"
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

func Ready(mgr *manager.Manager, group string, name string, dependsOn []v1.DependsOn) (bool, error) {
	if len(dependsOn) > 0 {
		var allDependenciesSolved = true
		logger.Log.Info("trying to solve dependencies", zap.String("group", group), zap.String("name", name))

		c := make(chan State)
		for _, depend := range dependsOn {
			dependency := NewDependencyFromDefinition(depend)
			go SolveDepends(mgr, dependency, c)
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
						go SolveDepends(mgr, d.Depend, c)
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
			mgr.Registry.Containers[group][name].Status.TransitionState(status.STATUS_DEPENDS_FAILED)
			return false, errors.New("didn't solve all dependencies")
		} else {
			mgr.Registry.Containers[group][name].Status.TransitionState(status.STATUS_DEPENDS_SOLVED)
			logger.Log.Info("all dependencies solved", zap.String("group", group), zap.String("name", name))
			return true, nil
		}
	}

	logger.Log.Info("no dependencies defined", zap.String("group", group), zap.String("name", name))
	mgr.Registry.Containers[group][name].Status.TransitionState(status.STATUS_DEPENDS_SOLVED)

	return true, nil
}

func SolveDepends(mgr *manager.Manager, depend *Dependency, c chan State) {
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

		go Depends(mgr, depend, ch)

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

func Depends(mgr *manager.Manager, depend *Dependency, ch chan State) {
	group, id := utils.ExtractGroupAndId(depend.Name)

	logger.Log.Info("trying to check if depends solved", zap.String("group", group), zap.String("name", id))

	if mgr.Registry.Containers[group] != nil {
		if id == "*" {
			for _, container := range mgr.Registry.Containers[group] {
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
			if mgr.Registry.Containers[group][id] != nil {
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
