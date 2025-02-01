package readiness

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/cenkalti/backoff/v4"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/secrets"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/types"
	"github.com/simplecontainer/smr/pkg/kinds/container/status"
	"go.uber.org/zap"
	"net/http"
	"regexp"
	"strings"
	"time"
)

func Ready(client *client.Http, container platforms.IContainer, user *authentication.User, channel chan *ReadinessState, logger *zap.Logger) (bool, error) {
	for _, ready := range container.GetDefinition().(*v1.ContainerDefinition).Spec.Container.Readiness {
		readiness, err := NewReadinessFromDefinition(client, user, container, ready)

		if err != nil {
			return false, err
		}

		readiness.Function = func() error {
			err = SolveReadiness(client, user, container, logger, readiness, channel)

			if err != nil {
				logger.Info(err.Error())
			}

			return err
		}

		backOff := backoff.WithContext(backoff.NewExponentialBackOff(), readiness.Ctx)

		err = backoff.Retry(readiness.Function, backOff)
		if err != nil {
			channel <- &ReadinessState{
				State: FAILED,
			}

			return false, err
		}
	}

	channel <- &ReadinessState{
		State: SUCCESS,
	}

	return true, nil
}

func NewReadinessFromDefinition(client *client.Http, user *authentication.User, container platforms.IContainer, readiness v1.ContainerReadiness) (*Readiness, error) {
	if readiness.Timeout == "" {
		readiness.Timeout = "30s"
	}

	timeout, err := time.ParseDuration(readiness.Timeout)

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	for index, value := range readiness.Command {
		regexDetectBigBrackets := regexp.MustCompile(`{{([^{\n}]*)}}`)
		matches := regexDetectBigBrackets.FindAllStringSubmatch(value, -1)

		if len(matches) > 0 {
			format := f.NewFromString(matches[0][1])

			if format.IsValid() && format.GetKind() == "secret" {
				continue
			} else {
				runtimeValue, ok := container.GetRuntime().Configuration.Map.Load(format.GetGroup())

				if ok {
					readiness.Command[index] = strings.Replace(readiness.Command[index], matches[0][0], runtimeValue.(string), 1)
				}
			}
		}
	}

	var commandUnpack []string
	commandUnpack, err = secrets.UnpackSecretsReadiness(client, user, readiness.Command)

	if err != nil {
		cancel()
		return nil, err
	}

	return &Readiness{
		Name:    readiness.Name,
		URL:     readiness.URL,
		Command: commandUnpack,
		Timeout: readiness.Timeout,
		Ctx:     ctx,
		Cancel:  cancel,
	}, nil
}

func SolveReadiness(client *client.Http, user *authentication.User, container platforms.IContainer, logger *zap.Logger, readiness *Readiness, channel chan *ReadinessState) error {
	if !container.GetStatus().IfStateIs(status.STATUS_READINESS_CHECKING) {
		readiness.Cancel()
	}

	channel <- &ReadinessState{
		State: CHECKING,
	}

	if readiness.URL != "" {
		readiness.Type = TYPE_URL
	}

	if len(readiness.Command) > 0 {
		readiness.Type = TYPE_COMMAND
	}

	switch readiness.Type {
	case TYPE_URL:
		jsonBytes, err := json.Marshal(readiness.BodyUnpack)

		logger.Info("readiness probe", zap.String("URL", readiness.URL), zap.String("data", string(jsonBytes)))

		var req *http.Request

		req, err = http.NewRequest("POST", readiness.URL, bytes.NewBuffer(jsonBytes))

		if err != nil {
			return errors.New("readiness request creation failed")
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Get(user.Username).Http.Do(req)
		if err != nil {
			return errors.New("readiness request failed")
		} else {
			if resp.StatusCode == http.StatusOK {
				return nil
			} else {
				return errors.New("readiness request failed")
			}
		}
	case TYPE_COMMAND:
		c, err := container.GetContainerState()
		if err == nil && c.State == "running" {
			var result types.ExecResult
			result, err = container.Exec(readiness.Command)

			if err != nil {
				return err
			}

			if result.Exit == 0 {
				return nil
			} else {
				return errors.New("readiness request failed")
			}
		} else {
			return errors.New("readiness request failed - container not running")
		}
	default:
		return nil
	}
}
