package main

import (
	"encoding/json"
	"github.com/mitchellh/mapstructure"
	"github.com/simplecontainer/smr/implementations/network/network"
	"github.com/simplecontainer/smr/implementations/network/shared"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/httpcontract"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/objects"
	"go.uber.org/zap"
	"net/http"
)

func (implementation *Implementation) Start(mgr *manager.Manager) error {
	implementation.Shared.Manager = mgr
	implementation.Started = true

	implementation.Shared.Client = mgr.Http

	return nil
}

func (implementation *Implementation) GetShared() interface{} {
	return implementation.Shared
}

func (implementation *Implementation) Apply(user *authentication.User, jsonData []byte) (httpcontract.ResponseImplementation, error) {
	var networkDefinition v1.NetworkDefinition

	if err := json.Unmarshal(jsonData, &networkDefinition); err != nil {
		return httpcontract.ResponseImplementation{
			HttpStatus:       400,
			Explanation:      "invalid configuration sent: json is not valid",
			ErrorExplanation: "invalid configuration sent: json is not valid",
			Error:            true,
			Success:          false,
		}, err
	}

	valid, err := networkDefinition.Validate()

	if !valid {
		return httpcontract.ResponseImplementation{
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

	mapstructure.Decode(data["network"], &networkDefinition)

	var format *f.Format
	format = f.New("network", networkDefinition.Meta.Group, networkDefinition.Meta.Name, "object")

	obj := objects.New(implementation.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = networkDefinition.ToJsonString()

	logger.Log.Debug("server received network object", zap.String("definition", jsonStringFromRequest))

	if obj.Exists() {
		if obj.Diff(jsonStringFromRequest) {
			err = obj.Update(format, jsonStringFromRequest)

			if err != nil {
				return httpcontract.ResponseImplementation{
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
			return httpcontract.ResponseImplementation{
				HttpStatus:       http.StatusInternalServerError,
				Explanation:      "",
				ErrorExplanation: err.Error(),
				Error:            true,
				Success:          false,
			}, err
		}
	}

	var networkObj *network.Network

	if obj.ChangeDetected() || !obj.Exists() {
		networkObj = network.New(jsonData)
	} else {
		networkObj = network.New(obj.GetDefinitionByte())
	}

	err = networkObj.Create()

	if err != nil {
		return httpcontract.ResponseImplementation{
			HttpStatus:       http.StatusInternalServerError,
			Explanation:      "",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}, err
	}

	return httpcontract.ResponseImplementation{
		HttpStatus:       200,
		Explanation:      "everything went smoothly: good job!",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
	}, nil
}

func (implementation *Implementation) Compare(user *authentication.User, jsonData []byte) (httpcontract.ResponseImplementation, error) {
	var network v1.NetworkDefinition

	if err := json.Unmarshal(jsonData, &network); err != nil {
		return httpcontract.ResponseImplementation{
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

	mapstructure.Decode(data["spec"], &network)

	var format *f.Format

	format = f.New("network", network.Meta.Group, network.Meta.Name, "object")
	obj := objects.New(implementation.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = network.ToJsonString()

	if obj.Exists() {
		obj.Diff(jsonStringFromRequest)
	} else {
		return httpcontract.ResponseImplementation{
			HttpStatus:       418,
			Explanation:      "object is drifted from the definition",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
		}, nil
	}

	if obj.ChangeDetected() {
		return httpcontract.ResponseImplementation{
			HttpStatus:       418,
			Explanation:      "object is drifted from the definition",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
		}, nil
	} else {
		return httpcontract.ResponseImplementation{
			HttpStatus:       200,
			Explanation:      "object in sync",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
		}, nil
	}
}

func (implementation *Implementation) Delete(user *authentication.User, jsonData []byte) (httpcontract.ResponseImplementation, error) {
	var network v1.NetworkDefinition

	if err := json.Unmarshal(jsonData, &network); err != nil {
		return httpcontract.ResponseImplementation{
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

	mapstructure.Decode(data["network"], &network)

	format := f.New("network", network.Meta.Group, network.Meta.Name, "object")

	obj := objects.New(implementation.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	if obj.Exists() {
		deleted, err := obj.Remove(format)

		if deleted {
			format = f.New("network", network.Meta.Group, network.Meta.Name, "")
			deleted, err = obj.Remove(format)

			return httpcontract.ResponseImplementation{
				HttpStatus:       200,
				Explanation:      "deleted resource successfully",
				ErrorExplanation: "",
				Error:            true,
				Success:          false,
			}, err
		} else {
			return httpcontract.ResponseImplementation{
				HttpStatus:       500,
				Explanation:      "failed to delete resource",
				ErrorExplanation: err.Error(),
				Error:            true,
				Success:          false,
			}, err
		}
	} else {
		return httpcontract.ResponseImplementation{
			HttpStatus:       404,
			Explanation:      "object not found",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
		}, err
	}
}

var Network Implementation = Implementation{
	Started: false,
	Shared:  &shared.Shared{},
}
