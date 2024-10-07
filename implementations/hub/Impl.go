package main

import (
	"github.com/simplecontainer/smr/implementations/hub/hub"
	"github.com/simplecontainer/smr/implementations/hub/shared"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/httpcontract"
	"github.com/simplecontainer/smr/pkg/manager"
)

func (implementation *Implementation) Start(mgr *manager.Manager) error {
	implementation.Shared.Manager = mgr
	implementation.Started = true

	implementation.Shared.Event = make(chan *hub.Event)

	return nil
}

func (implementation *Implementation) Apply(user *authentication.User, jsonData []byte) (httpcontract.ResponseImplementation, error) {
	return httpcontract.ResponseImplementation{
		HttpStatus:       501,
		Explanation:      "not implemented",
		ErrorExplanation: "",
		Error:            true,
		Success:          false,
	}, nil
}

func (implementation *Implementation) Compare(user *authentication.User, jsonData []byte) (httpcontract.ResponseImplementation, error) {
	return httpcontract.ResponseImplementation{
		HttpStatus:       501,
		Explanation:      "not implemented",
		ErrorExplanation: "",
		Error:            true,
		Success:          false,
	}, nil
}

func (implementation *Implementation) Delete(user *authentication.User, jsonData []byte) (httpcontract.ResponseImplementation, error) {
	return httpcontract.ResponseImplementation{
		HttpStatus:       501,
		Explanation:      "not implemented",
		ErrorExplanation: "",
		Error:            true,
		Success:          false,
	}, nil
}

func (implementation *Implementation) GetShared() interface{} {
	return implementation.Shared
}

var Hub Implementation = Implementation{
	Started: false,
	Shared:  &shared.Shared{},
}
