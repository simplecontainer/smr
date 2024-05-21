package dependency

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/qdnqn/smr/pkg/definitions/v1"
	"github.com/qdnqn/smr/pkg/logger"
	"github.com/qdnqn/smr/pkg/manager"
	"github.com/qdnqn/smr/pkg/template"
	"github.com/qdnqn/smr/pkg/utils"
	"go.uber.org/zap"
	"net/http"
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
		Name:     depend.Name,
		Operator: depend.Operator,
		Timeout:  depend.Timeout,
		Body:     depend.Body,
		Solved:   depend.Solved,
		Ctx:      ctx,
	}
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

		go Depends(mgr, "https://smr-agent:8080/operators", depend, ch)

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

		mgr.Registry.Containers[group][name].Status.DependsSolved = true

		if !allDependenciesSolved {
			mgr.Registry.Containers[group][name].Status.DependsSolved = false
			return false, errors.New("didn't solve all dependencies")
		} else {
			logger.Log.Info("all dependencies solved", zap.String("group", group), zap.String("name", name))
			return true, nil
		}
	}

	logger.Log.Info("no dependencies defined", zap.String("group", group), zap.String("name", name))
	mgr.Registry.Containers[group][name].Status.DependsSolved = true

	return true, nil
}

func Depends(mgr *manager.Manager, host string, depend *Dependency, ch chan State) {
	if depend.Operator != "" {
		var err error
		jsonData, _, err := template.ParseTemplate(mgr.Badger, depend.Body, nil)

		group, _ := utils.ExtractGroupAndId(depend.Name)
		URL := fmt.Sprintf("%s/%s/%s", host, group, depend.Operator)

		jsonBytes, err := json.Marshal(jsonData)

		var req *http.Request

		req, err = http.NewRequest("POST", URL, bytes.NewBuffer(jsonBytes))
		req.Header.Set("Content-Type", "application/json")

		logger.Log.Info(fmt.Sprintf("trying to call operator: %s", URL))
		client, err := mgr.Keys.GenerateHttpClient()
		if err != nil {
			logger.Log.Error(err.Error())
		}

		resp, err := client.Do(req)
		if err != nil {
			logger.Log.Error(err.Error())
		}

		if resp.StatusCode == http.StatusOK {
			ch <- State{
				Success: true,
				Depend:  depend,
			}
		} else {
			ch <- State{
				Success: false,
				Depend:  depend,
			}
		}
	} else {
		group, id := utils.ExtractGroupAndId(depend.Name)

		logger.Log.Info("trying to check if depends solved", zap.String("group", group), zap.String("name", id))

		if mgr.Registry.Containers[group] != nil {
			if id == "*" {
				for _, container := range mgr.Registry.Containers[group] {
					if !container.Status.DependsSolved {
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
}
