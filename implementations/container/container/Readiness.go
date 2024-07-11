package container

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/simplecontainer/smr/implementations/container/status"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/static"
	"net/http"
	"time"
)

func NewReadinessFromDefinition(client *http.Client, readiness v1.Readiness) (*Readiness, error) {
	if readiness.Timeout == "" {
		readiness.Timeout = "30s"
	}

	timeout, err := time.ParseDuration(readiness.Timeout)

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	bodyUnpack, err := UnpackSecretsReadiness(client, readiness.Body)

	if err != nil {
		cancel()
		return nil, err
	}

	return &Readiness{
		Name:       readiness.Name,
		Operator:   readiness.Operator,
		Timeout:    readiness.Timeout,
		Body:       readiness.Body,
		BodyUnpack: bodyUnpack,
		Ctx:        ctx,
		Cancel:     cancel,
	}, nil
}

func (container *Container) Ready(client *http.Client) (bool, error) {
	container.Status.TransitionState(container.Static.GeneratedName, status.STATUS_READINESS)

	for _, ready := range container.Static.Definition.Spec.Container.Readiness {
		readiness, err := NewReadinessFromDefinition(client, ready)

		if err != nil {
			return false, err
		}

		readiness.Function = func() error {
			return SolveReadiness(client, container, readiness)
		}

		backOff := backoff.WithContext(backoff.NewExponentialBackOff(), readiness.Ctx)

		err = backoff.Retry(readiness.Function, backOff)
		if err != nil {
			container.Status.TransitionState(container.Static.GeneratedName, status.STATUS_READINESS_FAILED)

			return false, err
		}
	}

	container.Status.TransitionState(container.Static.GeneratedName, status.STATUS_READY)
	container.Status.LastReadiness = true
	container.Status.LastReadinessTimestamp = time.Now()

	return true, nil
}

func SolveReadiness(client *http.Client, container *Container, readiness *Readiness) error {
	if _, err := container.Get(); err != nil {
		readiness.Cancel()
	}

	format := f.NewFromString(readiness.Name)
	URL := fmt.Sprintf("%s/%s/%s", "https://%s/api/v1/operators", static.SMR_AGENT_URL, format.Kind, readiness.Operator)

	jsonBytes, err := json.Marshal(readiness.BodyUnpack)

	var req *http.Request

	req, err = http.NewRequest("POST", URL, bytes.NewBuffer(jsonBytes))

	if err != nil {
		return errors.New("readiness request creation failed")
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return errors.New("readiness request failed")
	} else {
		if resp.StatusCode == http.StatusOK {
			return nil
		} else {
			return errors.New("readiness request failed")
		}
	}
}
