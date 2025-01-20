package container

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/distributed"
	"github.com/simplecontainer/smr/pkg/events"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/static"
	"net/http"
)

var supportedControlOperations = []string{"List", "Get", "Remove", "View", "Restart"}

func (container *Container) ListSupported(request contracts.Control) contracts.Response {
	return common.Response(http.StatusOK, "", nil, network.ToJson(supportedControlOperations))
}

func (container *Container) List(request contracts.Control) contracts.Response {
	data := make(map[string]any)

	format := f.New(KIND, "", "", "")

	obj := objects.New(container.Shared.Client.Get(request.User.Username), request.User)
	objs, err := obj.FindMany(format)

	if err != nil {
		return common.Response(http.StatusInternalServerError, static.STATUS_RESPONSE_INTERNAL_ERROR, err, nil)
	}

	for k, v := range objs {
		data[k] = v.GetDefinition()
	}

	return common.Response(http.StatusOK, "", nil, network.ToJson(data))
}
func (container *Container) Get(request contracts.Control) contracts.Response {
	format := f.NewFromString(fmt.Sprintf("%s.%s.%s.%s", KIND, request.Group, request.Name, "object"))

	obj := objects.New(container.Shared.Client.Get(request.User.Username), request.User)
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
func (container *Container) View(request contracts.Control) contracts.Response {
	containerObj := container.Shared.Registry.Find(fmt.Sprintf("%s", request.Group), fmt.Sprintf("%s", request.Name))

	if containerObj == nil {
		return common.Response(http.StatusNotFound, static.STATUS_RESPONSE_NOT_FOUND, errors.New("container not found"), nil)
	}

	var definition = make(map[string]any)
	definition[containerObj.GetGeneratedName()] = containerObj

	return common.Response(http.StatusOK, "", nil, network.ToJson(definition))
}
func (container *Container) Restart(request contracts.Control) contracts.Response {
	containerObj := container.Shared.Registry.Find(request.Group, request.Name)

	if containerObj == nil {
		return common.Response(http.StatusNotFound, static.STATUS_RESPONSE_NOT_FOUND, errors.New("container not found"), nil)
	}

	event := events.New(events.EVENT_RESTART, static.KIND_CONTAINER, containerObj.GetGroup(), containerObj.GetGeneratedName(), nil)

	bytes, err := event.ToJson()

	if err != nil {
		return common.Response(http.StatusInternalServerError, static.STATUS_RESPONSE_INTERNAL_ERROR, err, nil)
	}

	container.Shared.Manager.Replication.EventsC <- distributed.NewEncode(event.GetKey(), bytes, container.Shared.Manager.Config.KVStore.Node, static.CATEGORY_EVENT)

	return common.Response(http.StatusOK, static.STATUS_RESPONSE_RESTART, nil, nil)
}
func (container *Container) Remove(request contracts.Control) contracts.Response {
	containerObj := container.Shared.Registry.Find(fmt.Sprintf("%s", request.Group), fmt.Sprintf("%s", request.Name))

	if containerObj == nil {
		return common.Response(http.StatusNotFound, static.STATUS_RESPONSE_NOT_FOUND, errors.New("container not found"), nil)
	}

	event := events.New(events.EVENT_DELETE, static.KIND_CONTAINER, containerObj.GetGroup(), containerObj.GetGeneratedName(), nil)

	bytes, err := event.ToJson()

	if err != nil {
		container.Shared.Manager.Replication.EventsC <- distributed.NewEncode(event.GetKey(), bytes, container.Shared.Manager.Config.KVStore.Node, static.CATEGORY_EVENT)
	}

	return common.Response(http.StatusOK, static.STATUS_RESPONSE_DELETED, nil, nil)
}
