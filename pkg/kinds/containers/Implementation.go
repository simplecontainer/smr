package containers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/containers/reconcile"
	"github.com/simplecontainer/smr/pkg/kinds/containers/watcher"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/objects"
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
	containersDefinition := &v1.ContainersDefinition{}

	if err := json.Unmarshal(jsonData, &containersDefinition); err != nil {
		return contracts.Response{
			HttpStatus:       400,
			Explanation:      "invalid definition sent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}, err
	}

	valid, err := containersDefinition.Validate()

	if !valid {
		return contracts.Response{
			HttpStatus:       400,
			Explanation:      "invalid definition sent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}, err
	}

	data := make(map[string]interface{})
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		panic(err)
	}

	var format *f.Format
	format = f.New("containers", containersDefinition.Meta.Group, containersDefinition.Meta.Name, "object")

	obj := objects.New(containers.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = containersDefinition.ToJsonString()

	logger.Log.Debug("server received containers object", zap.String("definition", jsonStringFromRequest))

	if obj.Exists() {
		if obj.Diff(jsonStringFromRequest) {
			err = obj.Update(format, jsonStringFromRequest)

			if err != nil {
				return contracts.Response{
					HttpStatus:       http.StatusInternalServerError,
					Explanation:      "failed to update object",
					ErrorExplanation: err.Error(),
					Error:            false,
					Success:          true,
				}, err
			}
		}
	} else {
		err = obj.Add(format, jsonStringFromRequest)

		if err != nil {
			return contracts.Response{
				HttpStatus:       http.StatusInternalServerError,
				Explanation:      "failed to add object",
				ErrorExplanation: err.Error(),
				Error:            false,
				Success:          true,
			}, err
		}
	}

	GroupIdentifier := fmt.Sprintf("%s.%s", containersDefinition.Meta.Group, containersDefinition.Meta.Name)
	containersFromDefinition := containers.Shared.Watcher.Find(GroupIdentifier)

	if obj.Exists() {
		if obj.ChangeDetected() || containersFromDefinition == nil {
			if containersFromDefinition == nil {
				containersFromDefinition = reconcile.NewWatcher(*containersDefinition, containers.Shared.Manager)
				containersFromDefinition.Logger.Info("containers object created")

				go reconcile.HandleTickerAndEvents(containers.Shared, user, containersFromDefinition, agent)
			} else {
				containersFromDefinition.Definition = *containersDefinition
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
		containersFromDefinition = reconcile.NewWatcher(*containersDefinition, containers.Shared.Manager)
		containersFromDefinition.Logger.Info("containers object created")

		go reconcile.HandleTickerAndEvents(containers.Shared, user, containersFromDefinition, agent)

		containers.Shared.Watcher.AddOrUpdate(GroupIdentifier, containersFromDefinition)
		reconcile.Container(containers.Shared, user, containersFromDefinition, agent)
	}

	return contracts.Response{
		HttpStatus:       http.StatusOK,
		Explanation:      "everything went smoothly: good job!",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
	}, nil
}
func (containers *Containers) Compare(user *authentication.User, jsonData []byte) (contracts.Response, error) {
	containersDefinition := &v1.ContainersDefinition{}

	if err := json.Unmarshal(jsonData, &containersDefinition); err != nil {
		return contracts.Response{
			HttpStatus:       400,
			Explanation:      "invalid definition sent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}, err
	}

	data := make(map[string]interface{})
	err := json.Unmarshal(jsonData, &data)
	if err != nil {
		panic(err)
	}

	var format *f.Format
	format = f.New("containers", containersDefinition.Meta.Group, containersDefinition.Meta.Name, "object")

	obj := objects.New(containers.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = containersDefinition.ToJsonString()

	if obj.Exists() {
		obj.Diff(jsonStringFromRequest)
	} else {
		return contracts.Response{
			HttpStatus:       418,
			Explanation:      "object is drifted from the definition",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
		}, nil
	}

	if obj.ChangeDetected() {
		return contracts.Response{
			HttpStatus:       418,
			Explanation:      "object is drifted from the definition",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
		}, nil
	} else {
		return contracts.Response{
			HttpStatus:       200,
			Explanation:      "object in sync",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
		}, nil
	}
}
func (containers *Containers) Delete(user *authentication.User, jsonData []byte, agent string) (contracts.Response, error) {
	containersDefinition := &v1.ContainersDefinition{}

	if err := json.Unmarshal(jsonData, &containersDefinition); err != nil {
		return contracts.Response{
			HttpStatus:       400,
			Explanation:      "invalid definition sent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}, err
	}

	data := make(map[string]interface{})
	err := json.Unmarshal(jsonData, &data)
	if err != nil {
		panic(err)
	}

	var format *f.Format
	format = f.New("containers", containersDefinition.Meta.Group, containersDefinition.Meta.Name, "object")

	obj := objects.New(containers.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	if !obj.Exists() {
		return contracts.Response{
			HttpStatus:       404,
			Explanation:      "",
			ErrorExplanation: "object not found on the server",
			Error:            true,
			Success:          false,
		}, nil
	}

	_, err = obj.Remove(format)

	if err != nil {
		return contracts.Response{
			HttpStatus:       500,
			Explanation:      "object removal failed",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}, nil
	}

	GroupIdentifier := fmt.Sprintf("%s.%s", containersDefinition.Meta.Group, containersDefinition.Meta.Name)

	if containers.Shared.Watcher.Find(GroupIdentifier) != nil {
		containers.Shared.Watcher.Find(GroupIdentifier).Syncing = true
		containers.Shared.Watcher.Find(GroupIdentifier).Cancel()
	}

	for _, definition := range containersDefinition.Spec {
		format = f.New("container", definition.Meta.Group, definition.Meta.Name, "object")
		obj = objects.New(containers.Shared.Client.Get(user.Username), user)
		obj.Find(format)

		if obj.Exists() {
			def, _ := definition.ToJsonStringWithKind()
			go containers.Shared.Manager.KindsRegistry["container"].Delete(user, []byte(def), agent)
		}
	}

	return contracts.Response{
		HttpStatus:       200,
		Explanation:      "action completed successfully",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
	}, nil
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
