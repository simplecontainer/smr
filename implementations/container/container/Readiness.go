package container

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/implementations/container/status"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/logger"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func (container *Container) Ready(client *http.Client, err error) (bool, error) {
	if err != nil {
		logger.Log.Error("failed to generate mtls https client")
		return false, nil
	}

	readiness := make([]Readiness, 0)

	container.Status.TransitionState(container.Static.GeneratedName, status.STATUS_READINESS)

	if len(container.Static.Definition.Spec.Container.Readiness) > 0 {
		var allReadinessSolved = true
		logger.Log.Info("trying to solve readiness", zap.String("group", container.Static.Group), zap.String("name", container.Static.GeneratedName))

		c := make(chan ReadinessState)
		for k, readinessElem := range container.Static.Readiness {
			readiness = append(readiness, readinessElem)
			container.Static.Readiness[k].BodyUnpack = container.UnpackSecretsReadiness(client, readinessElem.Body)

			var timeout time.Duration
			timeout, err = time.ParseDuration(readinessElem.Timeout)

			if err != nil {
				timeout, err = time.ParseDuration("15s")
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)

			container.Static.Readiness[k].Ctx = ctx
			container.Static.Readiness[k].Cancel = cancel

			go container.SolveReadiness(client, &container.Static.Readiness[k], c)
		}

		for len(readiness) > 0 {
			select {
			case d := <-c:
				if container.Runtime.State == "running" && container.Status.IfStateIs(status.STATUS_READINESS) {
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
							container.Status.TransitionState(container.Static.GeneratedName, status.STATUS_READINESS_FAILED)
							logger.Log.Info("readiness deadline exceeded", zap.String("group", container.Static.Group), zap.String("name", container.Static.GeneratedName))
							allReadinessSolved = false

							for i, v := range readiness {
								if v.Name == d.Readiness.Name {
									readiness = append(readiness[:i], readiness[i+1:]...)
								}
							}
						}
					}
				} else {
					return false, errors.New("container is not in running and readiness state")
				}
			}
		}

		if !allReadinessSolved {
			container.Status.TransitionState(container.Static.GeneratedName, status.STATUS_READINESS_FAILED)
			return false, errors.New("didn't solve all readiness probes")
		} else {
			container.Status.TransitionState(container.Static.GeneratedName, status.STATUS_READY)
			logger.Log.Info("all readiness probes solved", zap.String("group", container.Static.Group), zap.String("name", container.Static.GeneratedName))
			return true, nil
		}
	}

	logger.Log.Info("no readiness defined - state is ready", zap.String("group", container.Static.Group), zap.String("name", container.Static.GeneratedName))
	container.Status.TransitionState(container.Static.GeneratedName, status.STATUS_READY)

	return true, nil
}

func (container *Container) SolveReadiness(client *http.Client, readiness *Readiness, c chan ReadinessState) {
	defer readiness.Cancel()

	ch := make(chan ReadinessState)
	defer close(ch)

	logger.Log.Info("trying to solve readiness", zap.String("name", readiness.Name))

	go container.IsReady(client, "https://smr-agent:1443/api/v1/operators", readiness, ch)

	for {
		select {
		case d := <-ch:
			c <- d
		case <-readiness.Ctx.Done():
			c <- ReadinessState{
				Success:   false,
				Missing:   false,
				Timeout:   true,
				Readiness: readiness,
			}
		}
	}
}

func (container *Container) IsReady(client *http.Client, host string, readiness *Readiness, ch chan ReadinessState) {
	var err error

	format := f.NewFromString(readiness.Name)
	URL := fmt.Sprintf("%s/%s/%s", host, format.Kind, readiness.Operator)

	jsonBytes, err := json.Marshal(readiness.BodyUnpack)

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
