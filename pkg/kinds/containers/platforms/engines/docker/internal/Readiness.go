package internal

import (
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/readiness"
)

type Readinesses struct {
	Readinesses []*readiness.Readiness
}

func NewReadinesses(readinesses []v1.ContainersReadiness) (*Readinesses, error) {
	ReadinessesObj := &Readinesses{
		Readinesses: make([]*readiness.Readiness, 0),
	}

	var err error
	for _, readiness := range readinesses {
		err = ReadinessesObj.Add(readiness)

		if err != nil {
			return nil, err
		}
	}

	return ReadinessesObj, nil
}

func NewReadiness(readinessDefinition v1.ContainersReadiness) (*readiness.Readiness, error) {
	r, err := readiness.NewReadinessFromDefinition(readinessDefinition)

	if err != nil {
		return nil, err
	}

	return r, nil
}

func (Readinesses *Readinesses) Add(Readiness v1.ContainersReadiness) error {
	r, err := NewReadiness(Readiness)

	if err != nil {
		return err
	}

	Readinesses.Readinesses = append(Readinesses.Readinesses, r)

	return nil
}
