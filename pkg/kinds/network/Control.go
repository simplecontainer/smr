package network

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	networkURL "github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/static"
	"net/http"
)

var supportedControlOperations = []string{"List", "Get", "Remove"}

func (network *Network) ListSupported(request contracts.Control) contracts.Response {
	return common.Response(http.StatusOK, "", nil, networkURL.ToJson(supportedControlOperations))
}

func (network *Network) List(request contracts.Control) contracts.Response {
	data := make(map[string]any)

	format := f.New(KIND, "", "", "")

	obj := objects.New(network.Shared.Client.Get(request.User.Username), request.User)
	objs, err := obj.FindMany(format)

	if err != nil {
		return common.Response(http.StatusInternalServerError, static.STATUS_RESPONSE_INTERNAL_ERROR, err, nil)
	}

	for k, v := range objs {
		data[k] = v.GetDefinition()
	}

	return common.Response(http.StatusOK, "", nil, networkURL.ToJson(data))
}

func (network *Network) Get(request contracts.Control) contracts.Response {
	format := f.NewFromString(fmt.Sprintf("%s.%s.%s.%s", KIND, request.Group, request.Name, "object"))

	obj := objects.New(network.Shared.Client.Get(request.User.Username), request.User)
	err := obj.Find(format)

	if err != nil {
		return common.Response(http.StatusNotFound, static.STATUS_RESPONSE_NOT_FOUND, err, nil)
	}

	definitionObject := obj.GetDefinition()

	var definition = make(map[string]any)
	definition["kind"] = KIND
	definition[KIND] = definitionObject

	return common.Response(http.StatusOK, "", nil, networkURL.ToJson(definition))
}

func (network *Network) Remove(request contracts.Control) contracts.Response {
	GroupIdentifier := fmt.Sprintf("%s.%s", request.Group, request.Name)
	format := f.NewFromString(GroupIdentifier)

	obj := objects.New(network.Shared.Client.Get(request.User.Username), request.User)
	err := obj.Find(format)

	if err != nil {
		return common.Response(http.StatusNotFound, static.STATUS_RESPONSE_NOT_FOUND, err, nil)
	}

	removed, err := obj.Remove(format)

	if !removed {
		return common.Response(http.StatusInternalServerError, static.STATUS_RESPONSE_INTERNAL_ERROR, err, nil)
	} else {
		return common.Response(http.StatusOK, static.STATUS_RESPONSE_DELETED, nil, nil)
	}
}
