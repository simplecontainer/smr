package main

import (
	"encoding/json"
	"errors"
	"github.com/mitchellh/mapstructure"
	"smr/pkg/database"
	"smr/pkg/definitions"
	"smr/pkg/implementations"
	"smr/pkg/manager"
	"smr/pkg/objects"
)

func (implementation *Implementation) Implementation(mgr *manager.Manager, jsonData []byte) (implementations.Response, error) {
	var certkey definitions.CertKey

	if err := json.Unmarshal(jsonData, &certkey); err != nil {
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

	mapstructure.Decode(data["certkey"], &certkey)

	var format database.FormatStructure

	format = database.Format("certkey", certkey.Meta.Group, certkey.Meta.Identifier, "object")
	obj := objects.New()
	err = obj.Find(mgr.Registry.Object, mgr.Badger, format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = certkey.ToJsonString()

	if obj.Exists() {
		if obj.Diff(jsonStringFromRequest) {
			err = obj.Update(mgr.Registry.Object, mgr.Badger, format, jsonStringFromRequest)
		}
	} else {
		err = obj.Add(mgr.Registry.Object, mgr.Badger, format, jsonStringFromRequest)
	}

	if obj.ChangeDetected() || !obj.Exists() {
		mgr.EmitChange(KIND, certkey.Meta.Group, certkey.Meta.Identifier)
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

var Certkey Implementation
