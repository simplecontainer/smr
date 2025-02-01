package custom

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

func (custom *Custom) Start() error {
	custom.Started = true
	return nil
}
func (custom *Custom) GetShared() interface{} {
	return custom.Shared
}

func (custom *Custom) Apply(user *authentication.User, jsonData []byte, agent string) (contracts.Response, error) {
	request, err := common.NewRequest(static.KIND_CERTKEY)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	if err = request.Definition.FromJson(jsonData); err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	definition := request.Definition.Definition.(*v1.CertKeyDefinition)

	valid, err := definition.Validate()

	if !valid {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	format := f.New(definition.GetPrefix(), static.CATEGORY_KIND, static.KIND_CERTKEY, definition.Meta.Group, definition.Meta.Name)
	obj := objects.New(custom.Shared.Client.Get(user.Username), user)

	var jsonStringFromRequest []byte
	jsonStringFromRequest, err = definition.ToJson()

	logger.Log.Debug("server received custom object", zap.String("definition", string(jsonStringFromRequest)))

	_, err = request.Definition.Apply(format, obj, static.KIND_CERTKEY)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	}

	return common.Response(http.StatusOK, "object applied", nil, nil), nil
}
func (custom *Custom) Compare(user *authentication.User, jsonData []byte) (contracts.Response, error) {
	request, err := common.NewRequest(static.KIND_CERTKEY)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	if err = request.Definition.FromJson(jsonData); err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	definition := request.Definition.Definition.(*v1.CertKeyDefinition)

	format := f.New(definition.GetPrefix(), static.CATEGORY_KIND, static.KIND_CERTKEY, definition.Meta.Group, definition.Meta.Name)
	obj := objects.New(custom.Shared.Client.Get(user.Username), user)

	changed, err := request.Definition.Changed(format, obj)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	}

	if changed {
		return common.Response(http.StatusTeapot, "object drifted", nil, nil), nil
	}

	return common.Response(http.StatusOK, "object in sync", nil, nil), nil
}
func (custom *Custom) Delete(user *authentication.User, jsonData []byte, agent string) (contracts.Response, error) {
	request, err := common.NewRequest(static.KIND_CERTKEY)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	if err = request.Definition.FromJson(jsonData); err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	definition := request.Definition.Definition.(*v1.CertKeyDefinition)

	format := f.New(definition.GetPrefix(), static.CATEGORY_KIND, static.KIND_CERTKEY, definition.Meta.Group, definition.Meta.Name)
	obj := objects.New(custom.Shared.Client.Get(user.Username), user)

	_, err = request.Definition.Delete(format, obj, static.KIND_CERTKEY)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	}

	return common.Response(http.StatusOK, "object in deleted", nil, nil), nil
}

func (custom *Custom) Event(event contracts.Event) error {
	return nil
}
