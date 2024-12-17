package httpauth

import (
	"encoding/json"
	"errors"
	"github.com/mitchellh/mapstructure"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/objects"
	"go.uber.org/zap"
	"net/http"
	"reflect"
	"strings"
)

func (httpauth *Httpauth) Start() error {
	httpauth.Started = true
	return nil
}
func (httpauth *Httpauth) GetShared() interface{} {
	return httpauth.Shared
}
func (httpauth *Httpauth) Apply(user *authentication.User, jsonData []byte, agent string) (contracts.Response, error) {
	if err := json.Unmarshal(jsonData, &httpauth.Definition); err != nil {
		return contracts.Response{
			HttpStatus:       400,
			Explanation:      "invalid configuration sent: json is not valid",
			ErrorExplanation: "invalid configuration sent: json is not valid",
			Error:            true,
			Success:          false,
		}, err
	}

	valid, err := httpauth.Definition.Validate()

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

	mapstructure.Decode(data, &httpauth)

	var format *f.Format

	format = f.New("httpauth", httpauth.Definition.Meta.Group, httpauth.Definition.Meta.Name, "object")
	obj := objects.New(httpauth.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = httpauth.Definition.ToJsonString()

	logger.Log.Debug("server received httpauth object", zap.String("definition", jsonStringFromRequest))

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
		//pl := plugins.GetPlugin(implementation.Shared.Manager.Config.OptRoot, "hub.so")
		//sharedHub := pl.GetShared().(*hubShared.Shared)
		//
		//sharedHub.Event <- &hub.Event{
		//	Kind:  KIND,
		//	Group: httpauth.Meta.Group,
		//	Name:  httpauth.Meta.Name,
		//	Data: "",
		//}
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
func (httpauth *Httpauth) Compare(user *authentication.User, jsonData []byte) (contracts.Response, error) {
	if err := json.Unmarshal(jsonData, &httpauth.Definition); err != nil {
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

	mapstructure.Decode(data["httpauth"], &httpauth)

	var format *f.Format

	format = f.New("httpauth", httpauth.Definition.Meta.Group, httpauth.Definition.Meta.Name, "object")
	obj := objects.New(httpauth.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = httpauth.Definition.ToJsonString()

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
func (httpauth *Httpauth) Delete(user *authentication.User, jsonData []byte, agent string) (contracts.Response, error) {
	if err := json.Unmarshal(jsonData, &httpauth.Definition); err != nil {
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

	mapstructure.Decode(data["httpauth"], &httpauth.Definition)

	format := f.New("httpauth", httpauth.Definition.Meta.Group, httpauth.Definition.Meta.Name, "object")

	obj := objects.New(httpauth.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	if obj.Exists() {
		deleted, err := obj.Remove(format)

		if deleted {
			format = f.New("httpauth", httpauth.Definition.Meta.Group, httpauth.Definition.Meta.Name, "")
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

func (httpauth *Httpauth) Run(operation string, request contracts.Control) contracts.Response {
	reflected := reflect.TypeOf(httpauth)
	reflectedValue := reflect.ValueOf(httpauth)

	for i := 0; i < reflected.NumMethod(); i++ {
		method := reflected.Method(i)

		if operation == method.Name {
			inputs := []reflect.Value{reflect.ValueOf(request)}
			returnValue := reflectedValue.MethodByName(strings.ToTitle(operation)).Call(inputs)

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
