package resource

import (
	"encoding/json"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mitchellh/mapstructure"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/types"
	"github.com/simplecontainer/smr/pkg/kinds/container/shared"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/objects"
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

func (resource *Resource) Apply(user *authentication.User, jsonData []byte, agent string) (contracts.Response, error) {
	if err := json.Unmarshal(jsonData, &resource.Definition); err != nil {
		return contracts.Response{
			HttpStatus:       400,
			Explanation:      "invalid configuration sent: json is not valid",
			ErrorExplanation: "invalid configuration sent: json is not valid",
			Error:            true,
			Success:          false,
		}, err
	}

	valid, err := resource.Definition.Validate()

	if !valid {
		return contracts.Response{
			HttpStatus:       400,
			Explanation:      "invalid definition sent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}, err
	}

	err = json.Unmarshal(jsonData, &resource)
	if err != nil {
		return contracts.Response{
			HttpStatus:       http.StatusBadRequest,
			Explanation:      "invalid configuration sent: json is not valid",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}, err
	}

	var format *f.Format

	format = f.New("resource", resource.Definition.Meta.Group, resource.Definition.Meta.Name, "object")
	obj := objects.New(resource.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = resource.Definition.ToJsonString()

	logger.Log.Debug("server received resource object", zap.String("definition", jsonStringFromRequest))

	if obj.Exists() {
		if obj.Diff(jsonStringFromRequest) {
			err = obj.Update(format, jsonStringFromRequest)

			if err != nil {
				return contracts.Response{
					HttpStatus:       http.StatusInternalServerError,
					Explanation:      "",
					ErrorExplanation: err.Error(),
					Error:            true,
					Success:          false,
				}, err
			}
		}
	} else {
		err = obj.Add(format, jsonStringFromRequest)

		if err != nil {
			return contracts.Response{
				HttpStatus:       http.StatusInternalServerError,
				Explanation:      "",
				ErrorExplanation: err.Error(),
				Error:            true,
				Success:          false,
			}, err
		}
	}

	if obj.ChangeDetected() || !obj.Exists() {
		resource.Shared.Manager.KindsRegistry["container"].GetShared().(*shared.Shared).Watcher.EventChannel <- &types.Events{
			Kind:    KIND,
			Group:   resource.Definition.Meta.Group,
			Name:    resource.Definition.Meta.Name,
			Message: "detected resource update, reconcile container",
		}
	} else {
		return contracts.Response{
			HttpStatus:       200,
			Explanation:      "object is same as the one on the server",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
		}, errors.New("object is same on the server")
	}

	return contracts.Response{
		HttpStatus:       200,
		Explanation:      "everything went smoothly: good job!",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
	}, nil
}

func (resource *Resource) Compare(user *authentication.User, jsonData []byte) (contracts.Response, error) {
	if err := json.Unmarshal(jsonData, &resource.Definition); err != nil {
		return contracts.Response{
			HttpStatus:       400,
			Explanation:      "invalid configuration sent: json is not valid",
			ErrorExplanation: "invalid configuration sent: json is not valid",
			Error:            true,
			Success:          false,
		}, err
	}

	data := make(map[string]interface{})
	err := json.Unmarshal(jsonData, &data)
	if err != nil {
		panic(err)
	}

	mapstructure.Decode(data["spec"], &resource)

	var format *f.Format

	format = f.New("resource", resource.Definition.Meta.Group, resource.Definition.Meta.Name, "object")
	obj := objects.New(resource.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = resource.Definition.ToJsonString()

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

func (resource *Resource) Delete(user *authentication.User, jsonData []byte, agent string) (contracts.Response, error) {
	if err := json.Unmarshal(jsonData, &resource.Definition); err != nil {
		return contracts.Response{
			HttpStatus:       400,
			Explanation:      "invalid configuration sent: json is not valid",
			ErrorExplanation: "invalid configuration sent: json is not valid",
			Error:            true,
			Success:          false,
		}, err
	}

	data := make(map[string]interface{})
	err := json.Unmarshal(jsonData, &data)
	if err != nil {
		panic(err)
	}

	mapstructure.Decode(data["spec"], &resource)

	format := f.New("resource", resource.Definition.Meta.Group, resource.Definition.Meta.Name, "object")

	obj := objects.New(resource.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	if obj.Exists() {
		deleted, err := obj.Remove(format)

		if deleted {
			format = f.New("httpauth", resource.Definition.Meta.Group, resource.Definition.Meta.Name, "")
			deleted, err = obj.Remove(format)

			return contracts.Response{
				HttpStatus:       200,
				Explanation:      "deleted resource successfully",
				ErrorExplanation: "",
				Error:            true,
				Success:          false,
			}, err
		} else {
			return contracts.Response{
				HttpStatus:       500,
				Explanation:      "failed to delete resource",
				ErrorExplanation: err.Error(),
				Error:            true,
				Success:          false,
			}, err
		}
	} else {
		return contracts.Response{
			HttpStatus:       404,
			Explanation:      "object not found",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
		}, err
	}
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
