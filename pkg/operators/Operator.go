package operators

import (
	"errors"
	"fmt"
	"github.com/imroc/req/v3"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"regexp"
	"smr/pkg/database"
	"smr/pkg/definitions"
	"smr/pkg/logger"
	"smr/pkg/manager"
	"smr/pkg/utils"
	"strings"
	"sync"
	"time"
)

type Result struct {
	Data string `json:"data"`
}

func SolveDepends(mgr *manager.Manager, depend definitions.DependsOn, c chan DependsState) {
	var dependsState DependsState

	fmt.Println(depend)

	dependsState = Depends(mgr, "http://smr-agent:8080/operators", depend.Name, depend.Operator, depend.Body)

	c <- dependsState
}

func Ready(mgr *manager.Manager, group string, name string, dependsOn []definitions.DependsOn) (bool, error) {
	if len(dependsOn) > 0 {
		var wg sync.WaitGroup

		waitGroupCount := len(dependsOn)
		wg.Add(waitGroupCount)

		logger.Log.Info("Wait group for dependencies", zap.Int("waitGroupCount", waitGroupCount))
		c := make(chan DependsState)

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
						logger.Log.Info("Success from dependency", zap.Int("waitGroupCount", waitGroupCount))

						if waitGroupCount > 0 {
							waitGroupCount -= 1
							wg.Done()
						}

						dependsOn[i].Solved = true
					} else {
						time.Sleep(5 * time.Second)
					}
				case <-time.After(timeout):
					logger.Log.Info("Success from dependency", zap.Int("waitGroupCount", waitGroupCount))

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

		logger.Log.Info("Ready finished", zap.String("group", group), zap.String("name", name), zap.Bool("DependsSolver", mgr.Registry.Containers[group][name].Status.DependsSolved))

		mgr.Registry.Containers[group][name].Status.DependsSolved = true

		for _, depend := range dependsOn {
			if !depend.Solved {
				mgr.Registry.Containers[group][name].Status.DependsSolved = false
				return false, errors.New("didn't solve all dependencies")
			}
		}
	} else {
		fmt.Println(group)
		fmt.Println(name)
		mgr.Registry.Containers[group][name].Status.DependsSolved = true
	}

	logger.Log.Info("Ready finished", zap.String("group", group), zap.String("name", name), zap.Bool("DependsSolver", mgr.Registry.Containers[group][name].Status.DependsSolved))

	return true, nil
}

func Depends(mgr *manager.Manager, host string, name string, operator string, json map[string]any) DependsState {
	client := req.C().DevMode()

	if operator != "" {
		for key, value := range json {
			// If {{ANYTHING_HERE}} is detected create template.Format type so that we can query the KV store if the format is valid
			format := database.FormatStructure{}

			regexDetectBigBrackets := regexp.MustCompile(`{([^{{}}]*)}`)
			matches := regexDetectBigBrackets.FindAllStringSubmatch(value.(string), -1)

			if len(matches) > 0 {
				SplitByDot := strings.SplitN(matches[0][1], ".", 3)

				regexExtractGroupAndId := regexp.MustCompile(`([^\[\]]*)`)
				GroupAndIdExtractor := regexExtractGroupAndId.FindAllStringSubmatch(SplitByDot[1], -1)

				if len(GroupAndIdExtractor) > 1 {
					format = database.Format(SplitByDot[0], GroupAndIdExtractor[0][0], GroupAndIdExtractor[1][0], SplitByDot[2])

					if format.Identifier != "*" {
						format.Identifier = fmt.Sprintf("%s-%s", viper.GetString("project"), GroupAndIdExtractor[1][0])
					}

					dbKey := strings.TrimSpace(fmt.Sprintf("%s.%s.%s.%s", format.Kind, format.Group, format.Identifier, format.Key))

					val, err := database.Get(mgr.Badger, dbKey)

					if err != nil {
						logger.Log.Error(err.Error())
					}

					json[key] = val
				}
			}
		}

		group, _ := utils.ExtractGroupAndId(name)
		url := fmt.Sprintf("%s/%s/%s", host, group, operator)
		var result Result

		logger.Log.Info(fmt.Sprintf("Trying to call operator: %s", url))
		resp, err := client.R().
			SetBody(&json).
			SetSuccessResult(&result).
			Post(url)
		if err != nil {
			logger.Log.Error(err.Error())
		}

		if resp.IsSuccessState() {
			return DependsState{
				Name:    fmt.Sprintf("%s/%s", name, operator),
				Success: true,
			}
		} else {
			return DependsState{
				Name:    fmt.Sprintf("%s/%s", name, operator),
				Success: false,
			}
		}
	} else {
		status := false

		group, id := utils.ExtractGroupAndId(name)

		logger.Log.Info("Trying to check if depends solved", zap.String("group", group), zap.String("name", id))

		if mgr.Registry.Containers[group] != nil {
			if id == "*" {
				status = true

				for _, container := range mgr.Registry.Containers[group] {
					if !container.Status.DependsSolved {
						return DependsState{
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

		return DependsState{
			Name:    fmt.Sprintf("%s/%s", name, operator),
			Success: status,
		}
	}
}
