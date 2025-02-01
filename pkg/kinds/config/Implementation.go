package config

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/static"
	"go.uber.org/zap"
	"net/http"
)

func (config *Config) Start() error {
	config.Started = true
	return nil
}
func (config *Config) GetShared() interface{} {
	return config.Shared
}

func (config *Config) Apply(user *authentication.User, jsonData []byte, agent string) (contracts.Response, error) {
	request, err := common.NewRequest(static.KIND_CONFIGURATION)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	if err = request.Definition.FromJson(jsonData); err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	definition := request.Definition.Definition.(*v1.ConfigurationDefinition)

	valid, err := definition.Validate()

	if !valid {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	format, _ := f.New(static.SMR_PREFIX, static.CATEGORY_KIND, static.KIND_CONFIGURATION, definition.Meta.Group, definition.Meta.Name)
	obj := objects.New(config.Shared.Client.Get(user.Username), user)

	err = obj.Find(format)

	var jsonStringFromRequest []byte
	jsonStringFromRequest, err = definition.ToJson()

	logger.Log.Debug("server received configuration object", zap.String("definition", string(jsonStringFromRequest)))

	obj, err = request.Definition.Apply(format, obj, static.KIND_CONFIGURATION)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	}

	if obj.ChangeDetected() {
		event := events.New(events.EVENT_CHANGE, static.KIND_CONTAINER, definition.GetKind(), definition.Meta.Group, definition.Meta.Name, nil)

		var bytes []byte
		bytes, err = event.ToJson()

		if err != nil {
			logger.Log.Debug("failed to dispatch event", zap.Error(err))
		} else {
			if config.Shared.Manager.Cluster.Node.NodeID == definition.GetRuntime().GetNode() {
				config.Shared.Manager.Cluster.KVStore.Propose(event.GetKey(), bytes, definition.GetRuntime().GetNode())
			}
		}
	}

	return common.Response(http.StatusOK, "object applied", nil, nil), nil
}
func (config *Config) Compare(user *authentication.User, jsonData []byte) (contracts.Response, error) {
	request, err := common.NewRequest(static.KIND_CONFIGURATION)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	if err = request.Definition.FromJson(jsonData); err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	definition := request.Definition.Definition.(*v1.ConfigurationDefinition)

	format, _ := f.New(static.SMR_PREFIX, static.CATEGORY_KIND, static.KIND_CONFIGURATION, definition.Meta.Group, definition.Meta.Name)
	obj := objects.New(config.Shared.Client.Get(user.Username), user)

	changed, err := request.Definition.Changed(format, obj)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	}

	if changed {
		return common.Response(http.StatusTeapot, "object drifted", nil, nil), nil
	}

	return common.Response(http.StatusOK, "object in sync", nil, nil), nil
}
func (config *Config) Delete(user *authentication.User, jsonData []byte, agent string) (contracts.Response, error) {
	request, err := common.NewRequest(static.KIND_CONFIGURATION)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	if err = request.Definition.FromJson(jsonData); err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	definition := request.Definition.Definition.(*v1.ConfigurationDefinition)

	format, _ := f.New(static.SMR_PREFIX, static.CATEGORY_KIND, static.KIND_CONFIGURATION, definition.Meta.Group, definition.Meta.Name)
	obj := objects.New(config.Shared.Client.Get(user.Username), user)

	_, err = request.Definition.Delete(format, obj, static.KIND_CONFIGURATION)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	}

	return common.Response(http.StatusOK, "object in deleted", nil, nil), nil
}

func (config *Config) Event(event contracts.Event) error {
	return nil
}
