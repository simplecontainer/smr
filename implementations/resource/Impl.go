package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/qdnqn/smr/implementations/resource/shared"
	"github.com/qdnqn/smr/pkg/database"
	"github.com/qdnqn/smr/pkg/definitions/v1"
	"github.com/qdnqn/smr/pkg/httpcontract"
	"github.com/qdnqn/smr/pkg/manager"
	"github.com/qdnqn/smr/pkg/objects"
	"github.com/spf13/viper"
)

func (implementation *Implementation) Start(mgr *manager.Manager) error {
	implementation.Shared.Manager = mgr
	implementation.Started = true

	return nil
}

func (implementation *Implementation) Apply(jsonData []byte) (httpcontract.ResponseImplementation, error) {
	var resource v1.Resource

	if err := json.Unmarshal(jsonData, &resource); err != nil {
		panic(err)
	}

	data := make(map[string]interface{})
	err := json.Unmarshal(jsonData, &data)
	if err != nil {
		return httpcontract.ResponseImplementation{
			HttpStatus:       400,
			Explanation:      "invalid resource sent: json is not valid",
			ErrorExplanation: "invalid resource sent: json is not valid",
			Error:            true,
			Success:          false,
		}, err
	}

	mapstructure.Decode(data["spec"], &resource)

	var format database.FormatStructure

	format = database.Format("resource", resource.Meta.Group, resource.Meta.Identifier, "object")
	obj := objects.New()
	err = obj.Find(implementation.Shared.Manager.Badger, format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = resource.ToJsonString()

	if obj.Exists() {
		if obj.Diff(jsonStringFromRequest) {
			err = obj.Update(implementation.Shared.Manager.Badger, format, jsonStringFromRequest)
		}
	} else {
		err = obj.Add(implementation.Shared.Manager.Badger, format, jsonStringFromRequest)
	}

	if obj.ChangeDetected() || !obj.Exists() {
		for key, value := range resource.Spec.Data {
			format = database.Format("resource", resource.Meta.Group, resource.Meta.Identifier, key)

			if format.Identifier != "*" {
				format.Identifier = fmt.Sprintf("%s-%s", viper.GetString("project"), resource.Meta.Identifier)
			}

			database.Put(implementation.Shared.Manager.Badger, format.ToString(), value.(string))
		}

		implementation.Shared.Manager.EmitChange(KIND, resource.Meta.Group, resource.Meta.Identifier)
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
	var resource v1.Resource

	if err := json.Unmarshal(jsonData, &resource); err != nil {
		panic(err)
	}

	data := make(map[string]interface{})
	err := json.Unmarshal(jsonData, &data)
	if err != nil {
		return httpcontract.ResponseImplementation{
			HttpStatus:       400,
			Explanation:      "invalid resource sent: json is not valid",
			ErrorExplanation: "invalid resource sent: json is not valid",
			Error:            true,
			Success:          false,
		}, err
	}

	mapstructure.Decode(data["resource"], &resource)

	var format database.FormatStructure

	format = database.Format("resource", resource.Meta.Group, resource.Meta.Identifier, "object")
	obj := objects.New()
	err = obj.Find(implementation.Shared.Manager.Badger, format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = resource.ToJsonString()

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
	var resource v1.Resource

	if err := json.Unmarshal(jsonData, &resource); err != nil {
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

	mapstructure.Decode(data["resource"], &resource)

	format := database.Format("resource", resource.Meta.Group, resource.Meta.Identifier, "object")

	obj := objects.New()
	err = obj.Find(implementation.Shared.Manager.Badger, format)

	if obj.Exists() {
		deleted, err := obj.Remove(implementation.Shared.Manager.Badger, format)

		if deleted {
			format = database.Format("resource", resource.Meta.Group, resource.Meta.Identifier, "")
			deleted, err = obj.Remove(implementation.Shared.Manager.Badger, format)

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

var Resource Implementation = Implementation{
	Shared:  &shared.Shared{},
	Started: false,
}
