package internal

import (
	"context"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
)

type Readinesses struct {
	Readinesses []*Readiness
}

type Readiness struct {
	Reference ReadinessReference
	Docker    ReadinessDocker
}

type ReadinessReference struct {
	Group      string
	Name       string
	Key        string
	MountPoint string
	Data       map[string]string
}

type ReadinessDocker struct {
	Name       string
	Operator   string
	Timeout    string
	Body       map[string]string
	Solved     bool
	BodyUnpack map[string]string  `json:"-"`
	Function   func() error       `json:"-"`
	Ctx        context.Context    `json:"-"`
	Cancel     context.CancelFunc `json:"-"`
}

func NewReadinesses(readinesses []v1.ContainerReadiness) *Readinesses {
	ReadinessesObj := &Readinesses{
		Readinesses: make([]*Readiness, 0),
	}

	for _, readiness := range readinesses {
		ReadinessesObj.Add(readiness)
	}

	return ReadinessesObj
}

func NewReadiness(readiness v1.ContainerReadiness) *Readiness {
	return &Readiness{
		Reference: ReadinessReference{},
		Docker:    ReadinessDocker{},
	}
}

func (Readinesses *Readinesses) Add(Readiness v1.ContainerReadiness) {
	Readinesses.Readinesses = append(Readinesses.Readinesses, NewReadiness(Readiness))
}
