package secret

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

func (secret *Secret) Start() error {
	secret.Started = true
	return nil
}
func (secret *Secret) GetShared() interface{} {
	return secret.Shared
}

func (secret *Secret) Apply(user *authentication.User, jsonData []byte, agent string) (contracts.Response, error) {
	request, err := common.NewRequest(static.KIND_SECRET)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	if err = request.Definition.FromJson(jsonData); err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	definition := request.Definition.Definition.(*v1.SecretDefinition)

	valid, err := definition.Validate()

	if !valid {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	format := f.New(definition.GetPrefix(), static.CATEGORY_KIND, static.KIND_SECRET, definition.Meta.Group, definition.Meta.Name)
	obj := objects.New(secret.Shared.Client.Get(user.Username), user)

	var jsonStringFromRequest []byte
	jsonStringFromRequest, err = definition.ToJson()

	logger.Log.Debug("server received secret object", zap.String("definition", string(jsonStringFromRequest)))

	_, err = request.Definition.Apply(format, obj, static.KIND_SECRET)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	}

	return common.Response(http.StatusOK, "object applied", nil, nil), nil
}
func (secret *Secret) Compare(user *authentication.User, jsonData []byte) (contracts.Response, error) {
	request, err := common.NewRequest(static.KIND_SECRET)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	if err = request.Definition.FromJson(jsonData); err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	definition := request.Definition.Definition.(*v1.SecretDefinition)

	format := f.New(definition.GetPrefix(), static.CATEGORY_KIND, static.KIND_SECRET, definition.Meta.Group, definition.Meta.Name)
	obj := objects.New(secret.Shared.Client.Get(user.Username), user)

	changed, err := request.Definition.Changed(format, obj)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	}

	if changed {
		return common.Response(http.StatusTeapot, "object drifted", nil, nil), nil
	}

	return common.Response(http.StatusOK, "object in sync", nil, nil), nil
}
func (secret *Secret) Delete(user *authentication.User, jsonData []byte, agent string) (contracts.Response, error) {
	request, err := common.NewRequest(static.KIND_SECRET)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	if err = request.Definition.FromJson(jsonData); err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	definition := request.Definition.Definition.(*v1.SecretDefinition)

	format := f.New(definition.GetPrefix(), static.CATEGORY_KIND, static.KIND_SECRET, definition.Meta.Group, definition.Meta.Name)
	obj := objects.New(secret.Shared.Client.Get(user.Username), user)

	_, err = request.Definition.Delete(format, obj, static.KIND_SECRET)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	}

	return common.Response(http.StatusOK, "object in deleted", nil, nil), nil
}

func (secret *Secret) Event(event contracts.Event) error {
	return nil
}
