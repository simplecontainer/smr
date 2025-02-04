package network

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/network/implementation"
	"github.com/simplecontainer/smr/pkg/static"
	"net/http"
)

func (network *Network) Start() error {
	network.Started = true
	return nil
}

func (network *Network) GetShared() interface{} {
	return network.Shared
}

func (network *Network) Apply(user *authentication.User, definition []byte, agent string) (contracts.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_NETWORK, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	obj, err := request.Apply(network.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	}

	var networkObj *implementation.Network

	if obj.ChangeDetected() || !obj.Exists() {
		networkObj = implementation.New(definition)
	} else {
		networkObj = implementation.New(obj.GetDefinitionByte())
	}

	err = networkObj.Create()

	if err != nil {
		return common.Response(http.StatusInternalServerError, "internal error", err, nil), err
	}

	return common.Response(http.StatusOK, "object applied", nil, nil), nil
}

func (network *Network) Compare(user *authentication.User, definition []byte) (contracts.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_NETWORK, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	_, err = request.Apply(network.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusTeapot, "object drifted", nil, nil), nil
	} else {
		return common.Response(http.StatusOK, "object in sync", nil, nil), nil
	}
}

func (network *Network) Delete(user *authentication.User, definition []byte, agent string) (contracts.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_NETWORK, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	_, err = request.Remove(network.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusInternalServerError, "", err, nil), err
	} else {
		return common.Response(http.StatusOK, "object deleted", nil, nil), nil
	}
}

func (network *Network) Event(event contracts.Event) error {
	return nil
}
