package resource

import (
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/distributed"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/events"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/static"
	"go.uber.org/zap"
	"net/http"
	"reflect"
	"strings"
)

func (resource *Resource) Start() error {
	resource.Started = true
	return nil
}

func (resource *Resource) GetShared() interface{} {
	return resource.Shared
}

func (resource *Resource) Propose(c *gin.Context, user *authentication.User, jsonData []byte, agent string) (contracts.Response, error) {
	request, err := common.NewRequest(static.KIND_RESOURCE)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	if err = request.Definition.FromJson(jsonData); err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	definition := request.Definition.Definition.(*v1.ResourceDefinition)

	valid, err := definition.Validate()

	if !valid {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	format := f.New("resource", definition.Meta.Group, definition.Meta.Name, "object")

	var bytes []byte
	bytes, err = definition.ToJsonWithKind()

	switch c.Request.Method {
	case http.MethodPost:
		resource.Shared.Manager.Cluster.KVStore.Propose(format.ToString(), bytes, static.CATEGORY_OBJECT, resource.Shared.Manager.Config.Node)
		break
	case http.MethodDelete:
		resource.Shared.Manager.Cluster.KVStore.Propose(format.ToString(), bytes, static.CATEGORY_OBJECT_DELETE, resource.Shared.Manager.Config.Node)
		break
	}

	return common.Response(http.StatusOK, "object applied", nil, nil), nil
}

func (resource *Resource) Apply(user *authentication.User, jsonData []byte, agent string) (contracts.Response, error) {
	request, err := common.NewRequest(static.KIND_RESOURCE)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	if err = request.Definition.FromJson(jsonData); err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	definition := request.Definition.Definition.(*v1.ResourceDefinition)

	valid, err := definition.Validate()

	if !valid {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	format := f.New("resource", definition.Meta.Group, definition.Meta.Name, "object")
	obj := objects.New(resource.Shared.Client.Get(user.Username), user)

	var jsonStringFromRequest []byte
	jsonStringFromRequest, err = definition.ToJson()

	logger.Log.Debug("server received resource object", zap.String("definition", string(jsonStringFromRequest)))

	obj, err = request.Definition.Apply(format, obj, static.KIND_RESOURCE)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	}

	if obj.ChangeDetected() {
		event := events.New(events.EVENT_CHANGE, definition.Meta.Group, definition.Meta.Name, nil)

		var bytes []byte
		bytes, err = event.ToJson()

		if err != nil {
			logger.Log.Debug("failed to dispatch event", zap.Error(err))
		} else {
			resource.Shared.Manager.Replication.EventsC <- distributed.NewEncode(event.GetKey(), bytes, agent, static.CATEGORY_EVENT)
		}

	}

	return common.Response(http.StatusOK, "object applied", nil, nil), nil
}

func (resource *Resource) Compare(user *authentication.User, jsonData []byte) (contracts.Response, error) {
	request, err := common.NewRequest(static.KIND_RESOURCE)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	if err = request.Definition.FromJson(jsonData); err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	definition := request.Definition.Definition.(*v1.ResourceDefinition)

	format := f.New("resource", definition.Meta.Group, definition.Meta.Name, "object")
	obj := objects.New(resource.Shared.Client.Get(user.Username), user)

	changed, err := request.Definition.Changed(format, obj)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	}

	if changed {
		return common.Response(http.StatusTeapot, "object drifted", nil, nil), nil
	}

	return common.Response(http.StatusOK, "object in sync", nil, nil), nil
}

func (resource *Resource) Delete(user *authentication.User, jsonData []byte, agent string) (contracts.Response, error) {
	request, err := common.NewRequest(static.KIND_RESOURCE)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	if err = request.Definition.FromJson(jsonData); err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	definition := request.Definition.Definition.(*v1.ResourceDefinition)

	format := f.New("resource", definition.Meta.Group, definition.Meta.Name, "object")
	obj := objects.New(resource.Shared.Client.Get(user.Username), user)

	_, err = request.Definition.Delete(format, obj, static.KIND_RESOURCE)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	}

	return common.Response(http.StatusOK, "object in deleted", nil, nil), nil
}

func (resource *Resource) Run(operation string, request contracts.Control) contracts.Response {
	reflected := reflect.TypeOf(resource)
	reflectedValue := reflect.ValueOf(resource)

	for i := 0; i < reflected.NumMethod(); i++ {
		method := reflected.Method(i)

		if operation == strings.ToLower(method.Name) {
			inputs := []reflect.Value{reflect.ValueOf(request)}
			returnValue := reflectedValue.MethodByName(method.Name).Call(inputs)

			return returnValue[0].Interface().(contracts.Response)
		}
	}

	return contracts.Response{
		HttpStatus:       400,
		Explanation:      "server doesn't support requested functionality",
		ErrorExplanation: "implementation is missing",
		Error:            true,
		Success:          false,
		Data:             nil,
	}
}
