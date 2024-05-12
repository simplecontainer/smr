package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/qdnqn/smr/pkg/database"
	"github.com/qdnqn/smr/pkg/definitions/v1"
	"github.com/qdnqn/smr/pkg/implementations"
	"github.com/qdnqn/smr/pkg/manager"
	"github.com/qdnqn/smr/pkg/objects"
	"github.com/spf13/viper"
)

func (implementation *Implementation) Implementation(mgr *manager.Manager, jsonData []byte) (implementations.Response, error) {
	var resource v1.Resource

	if err := json.Unmarshal(jsonData, &resource); err != nil {
		panic(err)
	}

	data := make(map[string]interface{})
	err := json.Unmarshal(jsonData, &data)
	if err != nil {
		return implementations.Response{
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
	err = obj.Find(mgr.Registry.Object, mgr.Badger, format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = resource.ToJsonString()

	if obj.Exists() {
		if obj.Diff(jsonStringFromRequest) {
			err = obj.Update(mgr.Registry.Object, mgr.Badger, format, jsonStringFromRequest)
		}
	} else {
		err = obj.Add(mgr.Registry.Object, mgr.Badger, format, jsonStringFromRequest)
	}

	if obj.ChangeDetected() || !obj.Exists() {
		for key, value := range resource.Spec.Data {
			format = database.Format("resource", resource.Meta.Group, resource.Meta.Identifier, key)

			if format.Identifier != "*" {
				format.Identifier = fmt.Sprintf("%s-%s", viper.GetString("project"), resource.Meta.Identifier)
			}

			database.Put(mgr.Badger, format.ToString(), value.(string))
		}

		mgr.EmitChange(KIND, resource.Meta.Group, resource.Meta.Identifier)
	} else {
		return implementations.Response{
			HttpStatus:       200,
			Explanation:      "object is same as the one on the server",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
		}, errors.New("object is same on the server")
	}

	return implementations.Response{
		HttpStatus:       200,
		Explanation:      "everything went smoothly: good job!",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
	}, nil
}

var Resource Implementation
