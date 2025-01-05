package certkey

import (
	"encoding/json"
	"errors"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/objects"
	"go.uber.org/zap"
	"net/http"
	"reflect"
	"strings"
)

func (certkey *Certkey) Start() error {
	certkey.Started = true
	return nil
}
func (certkey *Certkey) GetShared() interface{} {
	return certkey.Shared
}
func (certkey *Certkey) Apply(user *authentication.User, jsonData []byte, agent string) (contracts.Response, error) {
	var definition v1.CertKeyDefinition
	if err := json.Unmarshal(jsonData, &definition); err != nil {
		return contracts.Response{
			HttpStatus:       400,
			Explanation:      "invalid configuration sent: json is not valid",
			ErrorExplanation: "invalid configuration sent: json is not valid",
			Error:            true,
			Success:          false,
		}, err
	}

	valid, err := definition.Validate()

	if !valid {
		return contracts.Response{
			HttpStatus:       400,
			Explanation:      "invalid definition sent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}, err
	}

	var format *f.Format
	format = f.New("certkey", definition.Meta.Group, definition.Meta.Name, "object")

	obj := objects.New(certkey.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = definition.ToJsonString()

	logger.Log.Debug("server received certkey object", zap.String("definition", jsonStringFromRequest))

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
		//pl := plugins.GetPlugin(certkey.Shared.Manager.Config.OptRoot, "hub.so")
		//sharedHub := pl.GetShared().(*hubShared.Shared)
		//
		//sharedHub.Event <- &hub.Event{
		//	Kind:  KIND,
		//	Group: certkey.Definition.Meta.Group,
		//	Name:  certkey.Definition.Meta.Name,
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
func (certkey *Certkey) Compare(user *authentication.User, jsonData []byte) (contracts.Response, error) {
	var definition v1.CertKeyDefinition
	if err := json.Unmarshal(jsonData, &definition); err != nil {
		return contracts.Response{
			HttpStatus:       400,
			Explanation:      "invalid configuration sent: json is not valid",
			ErrorExplanation: "invalid configuration sent: json is not valid",
			Error:            true,
			Success:          false,
		}, err
	}

	var format *f.Format

	format = f.New("certkey", definition.Meta.Group, definition.Meta.Name, "object")
	obj := objects.New(certkey.Shared.Client.Get(user.Username), user)
	obj.Find(format)

	var jsonStringFromRequest string
	jsonStringFromRequest, _ = definition.ToJsonString()

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
func (certkey *Certkey) Delete(user *authentication.User, jsonData []byte, agent string) (contracts.Response, error) {
	var definition v1.CertKeyDefinition
	if err := json.Unmarshal(jsonData, &definition); err != nil {
		return contracts.Response{
			HttpStatus:       400,
			Explanation:      "invalid configuration sent: json is not valid",
			ErrorExplanation: "invalid configuration sent: json is not valid",
			Error:            true,
			Success:          false,
		}, err
	}

	format := f.New("certkey", definition.Meta.Group, definition.Meta.Name, "object")

	obj := objects.New(certkey.Shared.Client.Get(user.Username), user)
	obj.Find(format)

	if obj.Exists() {
		deleted, err := obj.Remove(format)

		if deleted {
			format = f.New("certkey", definition.Meta.Group, definition.Meta.Name, "")
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
		}, errors.New("object not found")
	}
}
func (certkey *Certkey) Run(operation string, request contracts.Control) contracts.Response {
	reflected := reflect.TypeOf(certkey)
	reflectedValue := reflect.ValueOf(certkey)

	for i := 0; i < reflected.NumMethod(); i++ {
		method := reflected.Method(i)

		if operation == strings.ToLower(method.Name) {
			inputs := []reflect.Value{reflect.ValueOf(request)}
			returnValue := reflectedValue.MethodByName(method.Name).Call(inputs)

			return returnValue[0].Interface().(contracts.Response)
		}
	}

	return contracts.Response{
		HttpStatus:       http.StatusBadRequest,
		Explanation:      "",
		ErrorExplanation: "server doesn't support requested functionality or permission suffice",
		Error:            true,
		Success:          false,
		Data:             nil,
	}
}
