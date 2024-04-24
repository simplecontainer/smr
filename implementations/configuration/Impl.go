package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"smr/pkg/database"
	"smr/pkg/definitions"
	"smr/pkg/implementations"
	"smr/pkg/manager"
	"smr/pkg/objects"
)

func (implementation *Implementation) Implementation(mgr *manager.Manager, jsonData []byte) (implementations.Response, error) {
	var config definitions.Configuration

	if err := json.Unmarshal(jsonData, &config); err != nil {
		return implementations.Response{
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

	var format database.FormatStructure

	format = database.Format("object-configuration", config.Meta.Group, config.Meta.Identifier, "object")
	obj := objects.New()
	err = obj.Find(mgr.Registry.Object, mgr.Badger, format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = config.ToJsonString()

	if obj.Exists() {
		if obj.Diff(jsonStringFromRequest) {
			err = obj.Update(mgr.Registry.Object, mgr.Badger, format, jsonStringFromRequest)
		}
	} else {
		err = obj.Add(mgr.Registry.Object, mgr.Badger, format, jsonStringFromRequest)
	}

	if obj.ChangeDetected() || !obj.Exists() {
		for key, value := range config.Spec.Data {
			format = database.Format("configuration", config.Meta.Group, config.Meta.Identifier, key)

			if format.Identifier != "*" {
				format.Identifier = fmt.Sprintf("%s-%s", viper.GetString("project"), config.Meta.Identifier)
			}

			database.Put(mgr.Badger, format.ToString(), value)
		}

		mgr.EmitChange(config.Meta.Group, config.Meta.Identifier)
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

var Configuration Implementation
