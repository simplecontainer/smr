package main

import "C"
import (
	"github.com/qdnqn/smr/implementations/container/container"
	"github.com/qdnqn/smr/implementations/container/shared"
	"github.com/qdnqn/smr/pkg/httpcontract"
	"github.com/qdnqn/smr/pkg/manager"
	"github.com/qdnqn/smr/pkg/registry"
	"strconv"
)

func (implementation *Implementation) Start(mgr *manager.Manager) error {
	implementation.Shared.Registry = &registry.Registry{
		Containers:     make(map[string]map[string]*container.Container),
		Indexes:        make(map[string][]int),
		BackOffTracker: make(map[string]map[string]int),
	}

	implementation.Shared.Manager = mgr
	implementation.Started = true

	return nil
}

func (implementation *Implementation) Apply(jsonData []byte) (httpcontract.ResponseImplementation, error) {
	implementation.State += 1

	return httpcontract.ResponseImplementation{
		HttpStatus:       200,
		Explanation:      strconv.Itoa(implementation.State),
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
	}, nil
}

func (implementation *Implementation) Compare(jsonData []byte) (httpcontract.ResponseImplementation, error) {
	return httpcontract.ResponseImplementation{
		HttpStatus:       200,
		Explanation:      "object in sync",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
	}, nil
}

func (implementation *Implementation) Delete(jsonData []byte) (httpcontract.ResponseImplementation, error) {
	return httpcontract.ResponseImplementation{
		HttpStatus:       200,
		Explanation:      "object in sync",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
	}, nil
}

// Exported
var Container Implementation = Implementation{
	Started: false,
	Shared:  &shared.Shared{},
}
