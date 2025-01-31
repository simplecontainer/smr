package containers

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/container/shared"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/static"
	"net/http"
)

var supportedControlOperations = []string{"List", "Get", "View"}

func (containers *Containers) ListSupported(request contracts.Control) contracts.Response {
	return common.Response(http.StatusOK, "", nil, network.ToJson(supportedControlOperations))
}

func (containers *Containers) List(request contracts.Control) contracts.Response {
	data := make(map[string]any)

	format := f.New(KIND, "", "", "")

	obj := objects.New(containers.Shared.Client.Get(request.User.Username), request.User)
	objs, err := obj.FindMany(format)

	if err != nil {
		return common.Response(http.StatusInternalServerError, static.STATUS_RESPONSE_INTERNAL_ERROR, err, nil)
	}

	for k, v := range objs {
		data[k] = v.GetDefinition()
	}

	return common.Response(http.StatusOK, "", nil, network.ToJson(data))
}
func (containers *Containers) Get(request contracts.Control) contracts.Response {
	format := f.NewFromString(fmt.Sprintf("/%s/%s/%s/%s", KIND, request.Group, request.Name, "object"))

	obj := objects.New(containers.Shared.Client.Get(request.User.Username), request.User)
	obj.Find(format)

	if !obj.Exists() {
		return common.Response(http.StatusNotFound, static.STATUS_RESPONSE_NOT_FOUND, errors.New(static.STATUS_RESPONSE_NOT_FOUND), nil)
	}

	r, err := common.NewRequest(KIND)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil)
	}

	err = r.Definition.FromJson(obj.GetDefinitionByte())

	if err != nil {
		return contracts.Response{}
	}

	bytes, err := r.Definition.ToJsonForUser()

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil)
	}

	return common.Response(http.StatusOK, "", nil, bytes)
}
func (containers *Containers) View(request contracts.Control) contracts.Response {
	registry := containers.Shared.Manager.KindsRegistry["container"].GetShared().(shared.Shared)
	container := registry.Registry.Find(fmt.Sprintf("%s", request.Group), fmt.Sprintf("%s", request.Name))

	if container == nil {
		return common.Response(http.StatusNotFound, static.STATUS_RESPONSE_NOT_FOUND, errors.New("container not found"), nil)
	}

	var definition = make(map[string]any)
	definition[container.GetGeneratedName()] = container

	return common.Response(http.StatusOK, "", nil, network.ToJson(definition))
}
