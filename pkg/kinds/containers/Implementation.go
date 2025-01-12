package containers

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/containers/reconcile"
	"github.com/simplecontainer/smr/pkg/kinds/containers/watcher"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/static"
	"go.uber.org/zap"
	"net/http"
	"reflect"
	"strings"
)

func (containers *Containers) Start() error {
	containers.Started = true

	containers.Shared.Watcher = &watcher.ContainersWatcher{}
	containers.Shared.Watcher.Containers = make(map[string]*watcher.Containers)

	return nil
}

func (containers *Containers) GetShared() interface{} {
	return containers.Shared
}
func (containers *Containers) Apply(user *authentication.User, jsonData []byte, agent string) (contracts.Response, error) {
	request, err := common.NewRequest(static.KIND_CONTAINERS)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	if err = request.Definition.FromJson(jsonData); err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	definition := request.Definition.Definition.(*v1.ContainersDefinition)

	valid, err := definition.Validate()

	if !valid {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	format := f.New("containers", definition.Meta.Group, definition.Meta.Name, "object")
	obj := objects.New(containers.Shared.Client.Get(user.Username), user)

	var jsonStringFromRequest []byte
	jsonStringFromRequest, err = definition.ToJson()

	logger.Log.Debug("server received containers object", zap.String("definition", string(jsonStringFromRequest)))

	obj, err = request.Definition.Apply(format, obj, static.KIND_CONTAINERS)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	}

	GroupIdentifier := fmt.Sprintf("%s.%s", definition.Meta.Group, definition.Meta.Name)
	containersFromDefinition := containers.Shared.Watcher.Find(GroupIdentifier)

	if obj.Exists() {
		if obj.ChangeDetected() || containersFromDefinition == nil {
			if containersFromDefinition == nil {
				containersFromDefinition = reconcile.NewWatcher(*definition, containers.Shared.Manager)
				containersFromDefinition.Logger.Info("containers object created")

				go reconcile.HandleTickerAndEvents(containers.Shared, user, containersFromDefinition, agent)
			} else {
				containersFromDefinition.Definition = *definition
				containersFromDefinition.Logger.Info("containers object modified")
			}

			containers.Shared.Watcher.AddOrUpdate(GroupIdentifier, containersFromDefinition)
			reconcile.Container(containers.Shared, user, containersFromDefinition, agent)
		} else {
			return contracts.Response{
				HttpStatus:       http.StatusOK,
				Explanation:      "containers object is same as the one on the server",
				ErrorExplanation: "",
				Error:            false,
				Success:          true,
			}, errors.New("containers object is same on the server")
		}
	} else {
		containersFromDefinition = reconcile.NewWatcher(*definition, containers.Shared.Manager)
		containersFromDefinition.Logger.Info("containers object created")

		go reconcile.HandleTickerAndEvents(containers.Shared, user, containersFromDefinition, agent)

		containers.Shared.Watcher.AddOrUpdate(GroupIdentifier, containersFromDefinition)
		reconcile.Container(containers.Shared, user, containersFromDefinition, agent)
	}

	return common.Response(http.StatusOK, "object applied", nil, nil), nil
}
func (containers *Containers) Compare(user *authentication.User, jsonData []byte) (contracts.Response, error) {
	request, err := common.NewRequest(static.KIND_CONTAINERS)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	if err = request.Definition.FromJson(jsonData); err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	definition := request.Definition.Definition.(*v1.ContainersDefinition)

	format := f.New("containers", definition.Meta.Group, definition.Meta.Name, "object")
	obj := objects.New(containers.Shared.Client.Get(user.Username), user)

	changed, err := request.Definition.Changed(format, obj)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	}

	if changed {
		return common.Response(http.StatusTeapot, "object drifted", nil, nil), nil
	}

	return common.Response(http.StatusOK, "object in sync", nil, nil), nil
}
func (containers *Containers) Delete(user *authentication.User, jsonData []byte, agent string) (contracts.Response, error) {
	request, err := common.NewRequest(static.KIND_CONTAINERS)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	if err = request.Definition.FromJson(jsonData); err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	definition := request.Definition.Definition.(*v1.ContainersDefinition)

	format := f.New("containers", definition.Meta.Group, definition.Meta.Name, "object")
	obj := objects.New(containers.Shared.Client.Get(user.Username), user)

	existingDefinition, err := request.Definition.Delete(format, obj, static.KIND_CONTAINERS)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	}

	GroupIdentifier := fmt.Sprintf("%s.%s", definition.Meta.Group, definition.Meta.Name)

	if containers.Shared.Watcher.Find(GroupIdentifier) != nil {
		containers.Shared.Watcher.Find(GroupIdentifier).Syncing = true
		containers.Shared.Watcher.Find(GroupIdentifier).Cancel()
	}

	for _, container := range existingDefinition.(*v1.ContainersDefinition).Spec {
		def, _ := container.ToJsonStringWithKind()
		go func() {
			_, err = containers.Shared.Manager.KindsRegistry["container"].Delete(user, []byte(def), agent)
			if err != nil {
				logger.Log.Error(err.Error())
			}
		}()
	}

	return common.Response(http.StatusOK, "object deleted", nil, nil), nil
}
func (containers *Containers) Run(operation string, request contracts.Control) contracts.Response {
	reflected := reflect.TypeOf(containers)
	reflectedValue := reflect.ValueOf(containers)

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
