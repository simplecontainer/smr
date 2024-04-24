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
	var config definitions.Configuration

	if err := json.Unmarshal(jsonData, &config); err != nil {
		panic(err)
	}

	data := make(map[string]interface{})
	err := json.Unmarshal(jsonData, &data)
	if err != nil {
		panic(err)
	}

	mapstructure.Decode(data["configuration"], &config)

	var format database.FormatStructure

	for key, value := range config.Spec.Data {
		format = database.Format("configuration", config.Meta.Group, config.Meta.Identifier, key)

		if format.Identifier != "*" {
			format.Identifier = fmt.Sprintf("%s-%s", viper.GetString("project"), config.Meta.Identifier)
		}

		dbKey := fmt.Sprintf("%s.%s.%s.%s", format.Kind, format.Group, format.Identifier, format.Key)
		database.Put(mgr.Badger, dbKey, value)
	}

	return KIND, nil
}

var Configuration ImplementationInternal
