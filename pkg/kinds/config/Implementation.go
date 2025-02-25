package config

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts/ievents"
	"github.com/simplecontainer/smr/pkg/contracts/iresponse"
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/static"
	"net/http"
)

func (config *Config) Start() error {
	config.Started = true
	return nil
}
func (config *Config) GetShared() interface{} {
	return config.Shared
}

func (config *Config) Apply(user *authentication.User, definition []byte, agent string) (iresponse.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_CONFIGURATION, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	obj, err := request.Apply(config.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	}

	if obj.ChangeDetected() {
		event := events.New(events.EVENT_CHANGE, static.KIND_CONFIGURATION, request.Definition.GetKind(), request.Definition.GetMeta().Group, request.Definition.GetMeta().Name, nil)

		if config.Shared.Manager.Cluster.Node.NodeID == request.Definition.GetRuntime().GetNode() {
			err = event.Propose(config.Shared.Manager.Cluster.KVStore, request.Definition.GetRuntime().GetNode())

			if err != nil {
				logger.Log.Error(err.Error())
			}
		}
	}

	return common.Response(http.StatusOK, "object applied", nil, nil), nil
}

func (config *Config) Delete(user *authentication.User, definition []byte, agent string) (iresponse.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_CONFIGURATION, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	_, err = request.Remove(config.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusTeapot, "", err, nil), err
	} else {
		return common.Response(http.StatusOK, "object deleted", nil, nil), nil
	}
}

func (config *Config) Event(event ievents.Event) error {
	return nil
}
