package main

import "C"
import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/qdnqn/smr/implementations/containers/reconcile"
	"github.com/qdnqn/smr/pkg/database"
	"github.com/qdnqn/smr/pkg/definitions/v1"
	"github.com/qdnqn/smr/pkg/httpcontract"
	"github.com/qdnqn/smr/pkg/manager"
	"github.com/qdnqn/smr/pkg/objects"
)

func (implementation *Implementation) Apply(mgr *manager.Manager, jsonData []byte, c *gin.Context) (httpcontract.ResponseImplementation, error) {
	containersDefinition := &v1.Containers{}

	if err := json.Unmarshal(jsonData, &containersDefinition); err != nil {
		return httpcontract.ResponseImplementation{
			HttpStatus:       400,
			Explanation:      "invalid definition sent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}, err
	}

	data := make(map[string]interface{})
	err := json.Unmarshal(jsonData, &data)
	if err != nil {
		panic(err)
	}

	var format database.FormatStructure
	format = database.Format("containers", containersDefinition.Meta.Group, containersDefinition.Meta.Name, "object")

	obj := objects.New()
	err = obj.Find(mgr.Badger, format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = containersDefinition.ToJsonString()

	if obj.Exists() {
		if obj.Diff(jsonStringFromRequest) {
			err = obj.Update(mgr.Badger, format, jsonStringFromRequest)
		}
	} else {
		err = obj.Add(mgr.Badger, format, jsonStringFromRequest)
	}

	if obj.ChangeDetected() || !obj.Exists() {
		containersFromDefinition := reconcile.NewWatcher(containersDefinition)
		GroupIdentifier := fmt.Sprintf("%s.%s", containersDefinition.Meta.Group, containersDefinition.Meta.Name)

		mgr.ContainersWatchers.AddOrUpdate(GroupIdentifier, containersFromDefinition)
		go reconcile.HandleTickerAndEvents(mgr, containersFromDefinition)
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

func (implementation *Implementation) Compare(mgr *manager.Manager, jsonData []byte, c *gin.Context) (httpcontract.ResponseImplementation, error) {
	definitionSent := &v1.Containers{}

	if err := json.Unmarshal(jsonData, &definitionSent); err != nil {
		return httpcontract.ResponseImplementation{
			HttpStatus:       400,
			Explanation:      "invalid definition sent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}, err
	} else {
		for _, definition := range definitionSent.Spec {
			format := database.Format("container", definition.Meta.Group, definition.Meta.Name, "object")
			obj := objects.New()
			err = obj.Find(mgr.Badger, format)

			var jsonStringFromRequest string
			jsonStringFromRequest, err = definition.ToJsonString()

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

		return httpcontract.ResponseImplementation{
			HttpStatus:       400,
			Explanation:      "definition is empty",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
		}, nil
	}
}

func (implementation *Implementation) Delete(mgr *manager.Manager, jsonData []byte, c *gin.Context) (httpcontract.ResponseImplementation, error) {
	return httpcontract.ResponseImplementation{
		HttpStatus:       404,
		Explanation:      "object not found on the server",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
	}, nil
}

// Exported
var Containers Implementation
