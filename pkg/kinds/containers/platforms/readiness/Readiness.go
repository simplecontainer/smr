package readiness

import (
	"context"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"time"
)

func NewReadinessFromDefinition(readiness v1.ContainersReadiness) (*Readiness, error) {
	if readiness.Timeout == "" {
		readiness.Timeout = "30s"
	}

	timeout, err := time.ParseDuration(readiness.Timeout)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	return &Readiness{
		Name:    readiness.Name,
		URL:     readiness.URL,
		Command: readiness.Command,
		Body:    readiness.Body,
		Timeout: readiness.Timeout,
		Ctx:     ctx,
		Cancel:  cancel,
	}, nil
}

func (r *Readiness) Reset() error {
	if r.Timeout == "" {
		r.Timeout = "30s"
	}

	timeout, err := time.ParseDuration(r.Timeout)

	if err != nil {
		return err
	}

	r.Ctx, r.Cancel = context.WithTimeout(context.Background(), timeout)
	return nil
}
