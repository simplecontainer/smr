package config

import (
	"encoding/json"
	"errors"
	"fmt"
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
)

func (config *Config) Start() error {
	config.Started = true
	return nil
}
func (config *Config) GetShared() interface{} {
	return config.Shared
}
func (config *Config) Apply(user *authentication.User, jsonData []byte, agent string) (contracts.Response, error) {
	if err := json.Unmarshal(jsonData, &config.Definition); err != nil {
		return contracts.Response{
			HttpStatus:       400,
			Explanation:      "invalid configuration sent: json is not valid",
			ErrorExplanation: "invalid configuration sent: json is not valid",
			Error:            true,
			Success:          false,
		}, err
	}

	valid, err := config.Definition.Validate()

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

	mapstructure.Decode(data, &config)

	var format *f.Format

	format = f.New("configuration", config.Definition.Meta.Group, config.Definition.Meta.Name, "object")
	obj := objects.New(config.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = config.Definition.ToJsonString()

	logger.Log.Debug("server received configuration object", zap.String("definition", jsonStringFromRequest))

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
		for key, _ := range config.Definition.Spec.Data {
			format = f.New("configuration", config.Definition.Meta.Group, config.Definition.Meta.Name, key)

			if format.Identifier != "*" {
				format.Identifier = fmt.Sprintf("%s-%s", config.Shared.Manager.Config.Environment.PROJECT, config.Definition.Meta.Name)
			}
		}

		config.Shared.Manager.KindsRegistry["container"].GetShared().(*shared.Shared).Watcher.EventChannel <- &types.Events{
			Kind:    KIND,
			Group:   config.Definition.Meta.Group,
			Name:    config.Definition.Meta.Name,
			Message: "detected configuration update, reconcile container",
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
func (config *Config) Compare(user *authentication.User, jsonData []byte) (contracts.Response, error) {
	if err := json.Unmarshal(jsonData, &config.Definition); err != nil {
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

	mapstructure.Decode(data["configuration"], &config)

	var format *f.Format

	format = f.New("configuration", config.Definition.Meta.Group, config.Definition.Meta.Name, "object")
	obj := objects.New(config.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = config.Definition.ToJsonString()

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
func (config *Config) Delete(user *authentication.User, jsonData []byte, agent string) (contracts.Response, error) {
	if err := json.Unmarshal(jsonData, &config.Definition); err != nil {
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

	mapstructure.Decode(data["configuration"], &config)

	format := f.New("configuration", config.Definition.Meta.Group, config.Definition.Meta.Name, "object")

	obj := objects.New(config.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	if obj.Exists() {
		deleted, err := obj.Remove(format)

		if deleted {
			format = f.New("configuration", config.Definition.Meta.Group, config.Definition.Meta.Name, "")
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
func (config *Config) Run(operation string, request contracts.Control) contracts.Response {
	reflected := reflect.TypeOf(config)
	reflectedValue := reflect.ValueOf(config)

	for i := 0; i < reflected.NumMethod(); i++ {
		method := reflected.Method(i)

		if operation == method.Name {
			inputs := []reflect.Value{reflect.ValueOf(request)}
			returnValue := reflectedValue.MethodByName(operation).Call(inputs)

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
