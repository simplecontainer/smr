package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/simplecontainer/smr/implementations/configuration/shared"
	"github.com/simplecontainer/smr/implementations/hub/hub"
	hubShared "github.com/simplecontainer/smr/implementations/hub/shared"
	"github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/httpcontract"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/plugins"
	"go.uber.org/zap"
)

func (implementation *Implementation) Start(mgr *manager.Manager) error {
	implementation.Shared.Manager = mgr
	implementation.Started = true

	client, err := manager.GenerateHttpClient(mgr.Keys)

	if err != nil {
		panic(err)
	}

	implementation.Client = client

	return nil
}

func (implementation *Implementation) GetShared() interface{} {
	return implementation.Shared
}

func (implementation *Implementation) Apply(jsonData []byte) (httpcontract.ResponseImplementation, error) {
	var config v1.Configuration

	if err := json.Unmarshal(jsonData, &config); err != nil {
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

	mapstructure.Decode(data["spec"], &config)

	var format objects.FormatStructure

	format = objects.Format("configuration", config.Meta.Group, config.Meta.Identifier, "object")
	obj := objects.New()
	err = obj.Find(implementation.Client, format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = config.ToJsonString()

	logger.Log.Debug("server received configuration object", zap.String("definition", jsonStringFromRequest))

	if obj.Exists() {
		if obj.Diff(jsonStringFromRequest) {
			err = obj.Update(implementation.Client, format, jsonStringFromRequest)
		}
	} else {
		err = obj.Add(implementation.Client, format, jsonStringFromRequest)
	}

	if obj.ChangeDetected() || !obj.Exists() {
		for key, value := range config.Spec.Data {
			format = objects.Format("configuration", config.Meta.Group, config.Meta.Identifier, key)

			if format.Identifier != "*" {
				format.Identifier = fmt.Sprintf("%s-%s", implementation.Shared.Manager.Config.Environment.PROJECT, config.Meta.Identifier)
			}

			obj.Update(implementation.Client, format, value)
		}

		pl := plugins.GetPlugin(implementation.Shared.Manager.Config.Root, "hub.so")
		sharedHub := pl.GetShared().(*hubShared.Shared)

		sharedHub.Event <- &hub.Event{
			Kind:       KIND,
			Group:      config.Meta.Group,
			Identifier: config.Meta.Identifier,
			Data:       nil,
		}
	} else {
		return httpcontract.ResponseImplementation{
			HttpStatus:       200,
			Explanation:      "object is same as the one on the server",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
		}, errors.New("object is same on the server")
	}

	return httpcontract.ResponseImplementation{
		HttpStatus:       200,
		Explanation:      "everything went smoothly: good job!",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
	}, nil
}

func (implementation *Implementation) Compare(jsonData []byte) (httpcontract.ResponseImplementation, error) {
	var config v1.Configuration

	if err := json.Unmarshal(jsonData, &config); err != nil {
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

	mapstructure.Decode(data["configuration"], &config)

	var format objects.FormatStructure

	format = objects.Format("configuration", config.Meta.Group, config.Meta.Identifier, "object")
	obj := objects.New()
	err = obj.Find(implementation.Client, format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = config.ToJsonString()

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

func (implementation *Implementation) Delete(jsonData []byte) (httpcontract.ResponseImplementation, error) {
	var config v1.Configuration

	if err := json.Unmarshal(jsonData, &config); err != nil {
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

	mapstructure.Decode(data["configuration"], &config)

	format := objects.Format("configuration", config.Meta.Group, config.Meta.Identifier, "object")

	obj := objects.New()
	err = obj.Find(implementation.Client, format)

	if obj.Exists() {
		deleted, err := obj.Remove(implementation.Client, format)

		if deleted {
			format = objects.Format("configuration", config.Meta.Group, config.Meta.Identifier, "")
			deleted, err = obj.Remove(implementation.Client, format)

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

var Configuration Implementation = Implementation{
	Started: false,
	Shared:  &shared.Shared{},
}
