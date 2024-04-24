package dependency

import (
	"errors"
	"fmt"
	"github.com/imroc/req/v3"
	"go.uber.org/zap"
	"smr/pkg/definitions"
	"smr/pkg/logger"
	"smr/pkg/manager"
	"smr/pkg/template"
	"smr/pkg/utils"
	"sync"
	"time"
)

func SolveDepends(mgr *manager.Manager, depend definitions.DependsOn, c chan State) {
	var State State

	State = Depends(mgr, "http://smr-agent:8080/operators", depend.Name, depend.Operator, depend.Body)

	c <- State
}

func Ready(mgr *manager.Manager, group string, name string, dependsOn []definitions.DependsOn) (bool, error) {
	if len(dependsOn) > 0 {
		var wg sync.WaitGroup

		waitGroupCount := len(dependsOn)
		wg.Add(waitGroupCount)

		logger.Log.Info("wait group for dependencies", zap.Int("waitGroupCount", waitGroupCount))
		c := make(chan State)

		for i, depend := range dependsOn {
			for {
				if dependsOn[i].Solved {
					break
				}

				timeout, err := time.ParseDuration(depend.Timeout)

				if err != nil {
					logger.Log.Error("invalid timeout provided")
					return false, errors.New("invalid timeout provided")
				}

				go SolveDepends(mgr, depend, c)

				select {
				case d := <-c:
					if d.Success {
						logger.Log.Info("success from dependency", zap.Int("waitGroupCount", waitGroupCount))

						if waitGroupCount > 0 {
							waitGroupCount -= 1
							wg.Done()
						}

						dependsOn[i].Solved = true
					} else {
						time.Sleep(5 * time.Second)
					}
				case <-time.After(timeout):
					logger.Log.Info("success from dependency", zap.Int("waitGroupCount", waitGroupCount))

					if waitGroupCount > 0 {
						waitGroupCount -= 1
						wg.Done()
					}

					dependsOn[i].Solved = false
				}
			}
		}
		wg.Wait()

		close(c)

		logger.Log.Info("ready finished", zap.String("group", group), zap.String("name", name))
		logger.Log.Info("updating status of dependency solver", zap.Bool("DependsSolver", mgr.Registry.Containers[group][name].Status.DependsSolved))

		mgr.Registry.Containers[group][name].Status.DependsSolved = true

		for _, depend := range dependsOn {
			if !depend.Solved {
				mgr.Registry.Containers[group][name].Status.DependsSolved = false
				return false, errors.New("didn't solve all dependencies")
			}
		}
	} else {
		mgr.Registry.Containers[group][name].Status.DependsSolved = true
	}

	logger.Log.Info("Ready finished", zap.String("group", group), zap.String("name", name), zap.Bool("DependsSolver", mgr.Registry.Containers[group][name].Status.DependsSolved))

	return true, nil
}

func Depends(mgr *manager.Manager, host string, name string, operator string, json map[string]any) State {
	client := req.C().DevMode()

	if operator != "" {
		var err error
		json, err = template.ParseTemplate(mgr.Badger, json)

		group, _ := utils.ExtractGroupAndId(name)
		url := fmt.Sprintf("%s/%s/%s", host, group, operator)
		var result Result

		logger.Log.Info(fmt.Sprintf("trying to call operator: %s", url))
		resp, err := client.R().
			SetBody(&json).
			SetSuccessResult(&result).
			Post(url)
		if err != nil {
			logger.Log.Error(err.Error())
		}

		if resp.IsSuccessState() {
			return State{
				Name:    fmt.Sprintf("%s/%s", name, operator),
				Success: true,
			}
		} else {
			return State{
				Name:    fmt.Sprintf("%s/%s", name, operator),
				Success: false,
			}
		}
	} else {
		status := false

		group, id := utils.ExtractGroupAndId(name)

		logger.Log.Info("trying to check if depends solved", zap.String("group", group), zap.String("name", id))

		if mgr.Registry.Containers[group] != nil {
			if id == "*" {
				status = true

				for _, container := range mgr.Registry.Containers[group] {
					if !container.Status.DependsSolved {
						return State{
							Name:    fmt.Sprintf("%s/%s", name, operator),
							Success: false,
						}
					}
				}
			} else {
				if mgr.Registry.Containers[group][id] != nil {
					status = mgr.Registry.Containers[group][id].Status.DependsSolved
				}
			}
		}

		return State{
			Name:    fmt.Sprintf("%s/%s", name, operator),
			Success: status,
		}
	}
}
