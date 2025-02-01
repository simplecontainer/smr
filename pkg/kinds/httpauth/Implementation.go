package httpauth

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/static"
	"go.uber.org/zap"
	"net/http"
)

func (httpauth *Httpauth) Start() error {
	httpauth.Started = true
	return nil
}
func (httpauth *Httpauth) GetShared() interface{} {
	return httpauth.Shared
}

func (httpauth *Httpauth) Apply(user *authentication.User, jsonData []byte, agent string) (contracts.Response, error) {
	request, err := common.NewRequest(static.KIND_HTTPAUTH)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	if err = request.Definition.FromJson(jsonData); err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	definition := request.Definition.Definition.(*v1.HttpAuthDefinition)

	valid, err := definition.Validate()

	if !valid {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	format, _ := f.New(static.SMR_PREFIX, static.CATEGORY_KIND, static.KIND_HTTPAUTH, definition.Meta.Group, definition.Meta.Name)
	obj := objects.New(httpauth.Shared.Client.Get(user.Username), user)

	var jsonStringFromRequest []byte
	jsonStringFromRequest, err = definition.ToJson()

	logger.Log.Debug("server received httpauth object", zap.String("definition", string(jsonStringFromRequest)))

	_, err = request.Definition.Apply(format, obj, static.KIND_HTTPAUTH)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	}

	return common.Response(http.StatusOK, "object applied", nil, nil), nil
}
func (httpauth *Httpauth) Compare(user *authentication.User, jsonData []byte) (contracts.Response, error) {
	request, err := common.NewRequest(static.KIND_HTTPAUTH)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	if err = request.Definition.FromJson(jsonData); err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	definition := request.Definition.Definition.(*v1.HttpAuthDefinition)

	format, _ := f.New(static.SMR_PREFIX, static.CATEGORY_KIND, static.KIND_HTTPAUTH, definition.Meta.Group, definition.Meta.Name)
	obj := objects.New(httpauth.Shared.Client.Get(user.Username), user)

	changed, err := request.Definition.Changed(format, obj)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	}

	if changed {
		return common.Response(http.StatusTeapot, "object drifted", nil, nil), nil
	}

	return common.Response(http.StatusOK, "object in sync", nil, nil), nil
}
func (httpauth *Httpauth) Delete(user *authentication.User, jsonData []byte, agent string) (contracts.Response, error) {
	request, err := common.NewRequest(static.KIND_HTTPAUTH)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	if err = request.Definition.FromJson(jsonData); err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	definition := request.Definition.Definition.(*v1.HttpAuthDefinition)

	format, _ := f.New(static.SMR_PREFIX, static.CATEGORY_KIND, static.KIND_HTTPAUTH, definition.Meta.Group, definition.Meta.Name)
	obj := objects.New(httpauth.Shared.Client.Get(user.Username), user)

	_, err = request.Definition.Delete(format, obj, static.KIND_HTTPAUTH)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	}

	return common.Response(http.StatusOK, "object in deleted", nil, nil), nil
}

func (httpauth *Httpauth) Event(event contracts.Event) error {
	return nil
}
