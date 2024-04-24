package main

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"smr/pkg/database"
	"smr/pkg/definitions"
	"smr/pkg/manager"
)

func (implementation *ImplementationInternal) ImplementationInternal(mgr *manager.Manager, jsonData []byte) (string, error) {
	var resource definitions.Resource

	if err := json.Unmarshal(jsonData, &resource); err != nil {
		panic(err)
	}

	data := make(map[string]interface{})
	err := json.Unmarshal(jsonData, &data)
	if err != nil {
		panic(err)
	}

	mapstructure.Decode(data["resource"], &resource)

	var format database.FormatStructure

	for key, value := range resource.Spec.Data {
		format = database.Format("resource", resource.Meta.Group, resource.Meta.Identifier, key)

		if format.Identifier != "*" {
			format.Identifier = fmt.Sprintf("%s-%s", viper.GetString("project"), resource.Meta.Identifier)
		}

		dbKey := fmt.Sprintf("%s.%s.%s.%s", format.Kind, format.Group, format.Identifier, format.Key)
		database.Put(mgr.Badger, dbKey, value)
	}

	return KIND, nil
}

var Resource ImplementationInternal
