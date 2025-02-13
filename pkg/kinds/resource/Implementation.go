package resource

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/static"
	"net/http"
)

func (resource *Resource) Start() error {
	resource.Started = true
	return nil
}

func (resource *Resource) GetShared() interface{} {
	return resource.Shared
}

func (resource *Resource) Apply(user *authentication.User, definition []byte, agent string) (contracts.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_RESOURCE, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	obj, err := request.Apply(resource.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	}

	if obj.ChangeDetected() {
		event := events.New(events.EVENT_CHANGE, static.KIND_CONTAINERS, request.Definition.GetKind(), request.Definition.GetMeta().Group, request.Definition.GetMeta().Name, nil)

		if resource.Shared.Manager.Cluster.Node.NodeID == request.Definition.GetRuntime().GetNode() {
			err = event.Propose(resource.Shared.Manager.Cluster.KVStore, request.Definition.GetRuntime().GetNode())

			if err != nil {
				logger.Log.Error(err.Error())
			}
		}
	}

	return common.Response(http.StatusOK, "object applied", nil, nil), nil
}

func (resource *Resource) Delete(user *authentication.User, definition []byte, agent string) (contracts.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_RESOURCE, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	_, err = request.Remove(resource.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusInternalServerError, "", err, nil), err
	} else {
		return common.Response(http.StatusOK, "object deleted", nil, nil), nil
	}
}

func (resource *Resource) Event(event contracts.Event) error {
	return nil
}
