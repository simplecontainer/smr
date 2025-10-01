package solver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/exec"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/readiness"
	"github.com/simplecontainer/smr/pkg/kinds/containers/status"
	"go.uber.org/zap"
	"net/http"
	"time"
)

var (
	ERROR_READINESS_INVALID_STATE  = errors.New("container is not in valid state for readiness checking")
	ERROR_CONTEXT_CANCELED         = errors.New("context canceled")
	ERROR_CONTEXT_TIMEOUT          = errors.New("context time out")
	ERROR_READINESS_RESET_FAILED   = errors.New("readiness reset failed")
	ERROR_INVALID_METHOD_FOR_URL   = errors.New("invalid method for url")
	ERROR_REQUEST_CREATION_FAILED  = errors.New("readiness request creation failed")
	ERROR_READINESS_REQUEST_FAILED = errors.New("readiness request failed")
	ERROR_COMMAND_FAILED           = errors.New("readiness command failed")
	ERROR_COMMAND_NOT_RUNNING      = errors.New("readiness command failed - container not running")
)

func Ready(ctx context.Context, client *clients.Http, container platforms.IContainer, user *authentication.User, channel chan *readiness.ReadinessState, logger *zap.Logger) (bool, error) {
	for _, r := range container.GetReadiness() {
		r.Function = func() error {
			select {
			case <-ctx.Done():
				state := &readiness.ReadinessState{
					State: readiness.CANCELED,
					Error: ERROR_CONTEXT_CANCELED,
				}
				channel <- state
				return backoff.Permanent(state.Error)

			case <-r.Ctx.Done():
				state := &readiness.ReadinessState{
					State: readiness.FAILED,
					Error: ERROR_CONTEXT_TIMEOUT,
				}
				channel <- state
				return backoff.Permanent(state.Error)

			default:
				container.GetStatus().LastReadinessStarted = time.Now()
				err := SolveReadiness(client, user, container, logger, r, channel)
				if err != nil {
					logger.Info(err.Error())
				}

				if errors.Is(err, ERROR_READINESS_INVALID_STATE) {
					return backoff.Permanent(err)
				}

				return err
			}
		}

		err := r.Reset()
		if err != nil {
			return false, ERROR_READINESS_RESET_FAILED
		}

		backOff := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
		err = backoff.Retry(r.Function, backOff)

		if ctx.Err() != nil {
			state := &readiness.ReadinessState{
				State: readiness.CANCELED,
				Error: ERROR_CONTEXT_CANCELED,
			}
			channel <- state
			return false, state.Error
		}

		if r.Ctx.Err() != nil {
			state := &readiness.ReadinessState{
				State: readiness.FAILED,
				Error: ERROR_CONTEXT_TIMEOUT,
			}
			channel <- state
			return false, state.Error
		}

		if err != nil {
			channel <- &readiness.ReadinessState{
				State: readiness.FAILED,
			}
			return false, err
		}
	}

	if ctx.Err() != nil {
		state := &readiness.ReadinessState{
			State: readiness.CANCELED,
			Error: ERROR_CONTEXT_CANCELED,
		}
		channel <- state
		return false, state.Error
	}

	select {
	case channel <- &readiness.ReadinessState{State: readiness.SUCCESS}:
	}

	return true, nil
}

func SolveReadiness(client *clients.Http, user *authentication.User, container platforms.IContainer, logger *zap.Logger, r *readiness.Readiness, channel chan *readiness.ReadinessState) error {
	if !container.GetStatus().IfStateIs(status.READINESS_CHECKING) && !container.GetStatus().IfStateIs(status.START) {
		logger.Info("container is not in valid state for readiness checking", zap.String("state", container.GetStatus().GetState()))
		return ERROR_READINESS_INVALID_STATE
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
		case "GET", "POST":
			// allowed
		default:
			return ERROR_INVALID_METHOD_FOR_URL
		}

		jsonBytes, err := json.Marshal(r.BodyUnpack)
		logger.Info("readiness probe", zap.String("URL", r.URL), zap.String("data", string(jsonBytes)))

		req, err := http.NewRequest(r.Method, r.URL, bytes.NewBuffer(jsonBytes))
		if err != nil {
			return ERROR_REQUEST_CREATION_FAILED
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Get(user.Username).Http.Do(req)
		if err != nil {
			return ERROR_READINESS_REQUEST_FAILED
		}
		if resp.StatusCode == http.StatusOK {
			r.Solved = true
			return nil
		}
		return ERROR_READINESS_REQUEST_FAILED

	case readiness.TYPE_COMMAND:
		c, err := container.GetState()
		if err == nil && c.State == "running" {
			session, err := exec.Create(r.Ctx, r.Cancel, nil, container, r.Command, false, "", "")
			if err != nil {
				return fmt.Errorf("readiness creating command failed: %w", err)
			}

			result, err := session.Output(container)
			if err != nil {
				return fmt.Errorf("readiness command failed: %w", err)
			}

			if result.Exit == 0 {
				r.Solved = true
				return nil
			}
			return ERROR_COMMAND_FAILED
		}
		return ERROR_COMMAND_NOT_RUNNING

	default:
		return nil
	}
}
