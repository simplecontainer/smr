package solver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/cenkalti/backoff/v4"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/readiness"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/types"
	"github.com/simplecontainer/smr/pkg/kinds/containers/status"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func Ready(ctx context.Context, client *client.Http, container platforms.IContainer, user *authentication.User, channel chan *readiness.ReadinessState, logger *zap.Logger) (bool, error) {
	for _, r := range container.GetReadiness() {
		r.Function = func() error {
			container.GetStatus().LastReadinessStarted = time.Now()
			err := SolveReadiness(client, user, container, logger, r, channel)

			if err != nil {
				logger.Info(err.Error())
			}

			return err
		}

		err := r.Reset()

		if err != nil {
			return false, errors.New("readiness reset failed")
		}

		backOff := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)

		err = backoff.Retry(r.Function, backOff)
		if err != nil {
			channel <- &readiness.ReadinessState{
				State: readiness.FAILED,
			}

			return false, err
		}
	}

	select {
	case <-ctx.Done():
		return false, ctx.Err() // avoid sending
	case channel <- &readiness.ReadinessState{State: readiness.SUCCESS}:
	}

	return true, nil
}

func SolveReadiness(client *client.Http, user *authentication.User, container platforms.IContainer, logger *zap.Logger, r *readiness.Readiness, channel chan *readiness.ReadinessState) error {
	if !container.GetStatus().IfStateIs(status.READINESS_CHECKING) {
		r.Cancel()
	}

	channel <- &readiness.ReadinessState{
		State: readiness.CHECKING,
	}

	if r.URL != "" {
		r.Type = readiness.TYPE_URL
	}

	if len(r.Command) > 0 {
		r.Type = readiness.TYPE_COMMAND
	}

	switch r.Type {
	case readiness.TYPE_URL:
		switch r.Method {
		case "GET":
		case "POST":
			break
		default:
			return errors.New("invalid method for url")
		}

		jsonBytes, err := json.Marshal(r.BodyUnpack)

		logger.Info("readiness probe", zap.String("URL", r.URL), zap.String("data", string(jsonBytes)))

		var req *http.Request

		req, err = http.NewRequest(r.Method, r.URL, bytes.NewBuffer(jsonBytes))

		if err != nil {
			return errors.New("readiness request creation failed")
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Get(user.Username).Http.Do(req)
		if err != nil {
			return errors.New("readiness request failed")
		} else {
			if resp.StatusCode == http.StatusOK {
				r.Solved = true
				return nil
			} else {
				return errors.New("readiness request failed")
			}
		}
	case readiness.TYPE_COMMAND:
		c, err := container.GetState()
		if err == nil && c.State == "running" {
			var result types.ExecResult
			result, err = container.Exec(r.Command)

			if err != nil {
				return err
			}

			if result.Exit == 0 {
				r.Solved = true
				return nil
			} else {
				return errors.New("readiness command failed")
			}
		} else {
			return errors.New("readiness command failed - container not running")
		}
	default:
		return nil
	}
}
