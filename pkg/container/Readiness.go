package container

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/qdnqn/smr/pkg/logger"
	"github.com/qdnqn/smr/pkg/utils"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func (container *Container) Ready(BadgerEncrypted *badger.DB, client *http.Client, err error) (bool, error) {
	if err != nil {
		logger.Log.Error("failed to genereate mtls https client")
		return false, nil
	}

	readiness := make([]Readiness, 0)

	if len(container.Static.Definition.Spec.Container.Readiness) > 0 {
		var allReadinessSolved = true
		logger.Log.Info("trying to solve readiness", zap.String("group", container.Static.Group), zap.String("name", container.Static.GeneratedName))

		c := make(chan ReadinessState)
		for _, readinessElem := range container.Static.Readiness {
			readiness = append(readiness, readinessElem)
			readinessElem.Body = container.UnpackSecretsReadiness(BadgerEncrypted, readinessElem.Body)
			go container.SolveReadiness(client, &readinessElem, c)
		}

		for len(readiness) > 0 {
			select {
			case d := <-c:
				if d.Missing {
					allReadinessSolved = false

					for i, v := range readiness {
						if v.Name == d.Readiness.Name {
							readiness = append(readiness[:i], readiness[i+1:]...)
						}
					}
				}

				if d.Success {
					logger.Log.Info("readiness solved", zap.String("group", container.Static.Group), zap.String("name", container.Static.GeneratedName))

					for i, v := range readiness {
						if v.Name == d.Readiness.Name {
							readiness = append(readiness[:i], readiness[i+1:]...)
						}
					}
				} else {
					deadline, _ := d.Readiness.Ctx.Deadline()

					if deadline.After(time.Now()) {
						time.Sleep(5 * time.Second)
						go container.SolveReadiness(client, d.Readiness, c)
					} else {
						logger.Log.Info("readiness deadline exceeded", zap.String("group", container.Static.Group), zap.String("name", container.Static.GeneratedName))
						allReadinessSolved = false

						for i, v := range readiness {
							if v.Name == d.Readiness.Name {
								readiness = append(readiness[:i], readiness[i+1:]...)
							}
						}
					}
				}
			}
		}

		if !allReadinessSolved {
			container.Status.Ready = false
			return false, errors.New("didn't solve all readiness probes")
		} else {
			container.Status.Ready = true
			logger.Log.Info("all readiness probes solved", zap.String("group", container.Static.Group), zap.String("name", container.Static.GeneratedName))
			return true, nil
		}
	}

	logger.Log.Info("no readiness defined - state is ready", zap.String("group", container.Static.Group), zap.String("name", container.Static.GeneratedName))
	container.Status.Ready = true

	return true, nil
}

func (container *Container) SolveReadiness(client *http.Client, readiness *Readiness, c chan ReadinessState) {
	if readiness.Timeout == "" {
		readiness.Timeout = "30s"
	}

	timeout, err := time.ParseDuration(readiness.Timeout)

	if err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		ch := make(chan ReadinessState)
		defer close(ch)

		logger.Log.Info("trying to solve readiness", zap.String("name", readiness.Name))

		go container.IsReady(client, "https://smr-agent:1443/api/v1/operators", readiness, ch)

		for {
			select {
			case d := <-ch:
				c <- d
			case <-ctx.Done():
				c <- ReadinessState{
					Success:   false,
					Missing:   false,
					Timeout:   true,
					Readiness: readiness,
				}
			}
		}
	} else {
		c <- ReadinessState{
			Success:   false,
			Missing:   false,
			Timeout:   false,
			Error:     err,
			Readiness: readiness,
		}
	}
}

func (container *Container) IsReady(client *http.Client, host string, readiness *Readiness, ch chan ReadinessState) {
	var err error

	group, _ := utils.ExtractGroupAndId(readiness.Name)
	URL := fmt.Sprintf("%s/%s/%s", host, group, readiness.Operator)

	jsonBytes, err := json.Marshal(readiness.Body)

	var req *http.Request

	req, err = http.NewRequest("POST", URL, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")

	logger.Log.Info(fmt.Sprintf("trying to call operator: %s", URL))

	resp, err := client.Do(req)
	if err != nil {
		logger.Log.Error(err.Error())

		ch <- ReadinessState{
			Success:   false,
			Readiness: readiness,
		}
	} else {
		if resp.StatusCode == http.StatusOK {
			ch <- ReadinessState{
				Success:   true,
				Readiness: readiness,
			}
		} else {
			ch <- ReadinessState{
				Success:   false,
				Readiness: readiness,
			}
		}
	}
}
