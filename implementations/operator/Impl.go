package main

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/r3labs/diff/v3"
	"go.uber.org/zap"
	"smr/pkg/database"
	"smr/pkg/definitions"
	"smr/pkg/implementations"
	"smr/pkg/logger"
	"smr/pkg/manager"
	"smr/pkg/objects"
	"smr/pkg/replicas"
)

func (implementation *Implementation) Implementation(mgr *manager.Manager, jsonData []byte) (implementations.Response, error) {
	var operatorContainer definitions.Container

	if err := json.Unmarshal(jsonData, &operatorContainer); err != nil {
		return implementations.Response{
			HttpStatus:       400,
			Explanation:      "invalid definition sent",
			ErrorExplanation: "json is not valid",
			Error:            true,
			Success:          false,
		}, err
	} else {
		data := make(map[string]interface{})
		err := json.Unmarshal(jsonData, &data)
		if err != nil {
			panic(err)
		}

		mapstructure.Decode(data["operator"], &operatorContainer)

		var globalGroups []string
		var globalNames []string

		format := database.Format("object-operator", operatorContainer.Meta.Group, operatorContainer.Meta.Name, "object")
		obj := objects.New()
		err = obj.Find(mgr.Registry.Object, mgr.Badger, format)

		var jsonStringFromRequest string
		jsonStringFromRequest, err = operatorContainer.ToJsonString()

		if obj.Exists() {
			if obj.Diff(jsonStringFromRequest) {
				// Detect only change on replicas, if that's true tackle only scale up or scale down without recreating
				// containers that are there already, otherwise recreate everything
				err = obj.Update(mgr.Registry.Object, mgr.Badger, format, jsonStringFromRequest)
			}
		} else {
			err = obj.Add(mgr.Registry.Object, mgr.Badger, format, jsonStringFromRequest)
		}

		if obj.ChangeDetected() || !obj.Exists() {
			logger.Log.Info("object is changed", zap.String("container", operatorContainer.Meta.Name))

			name := operatorContainer.Meta.Name
			logger.Log.Info(fmt.Sprintf("trying to generate container %s object", name))
			groups, names, err := implementation.generateContainerNameAndGroup(mgr, operatorContainer, obj.Changelog)

			if err == nil {
				logger.Log.Info(fmt.Sprintf("generated container %s object", name))

				globalGroups = append(globalGroups, groups...)
				globalNames = append(globalNames, names...)
			} else {
				logger.Log.Error("failed to generate names and groups")

				return implementations.Response{
					HttpStatus:       500,
					Explanation:      "failed to generate groups and names",
					ErrorExplanation: err.Error(),
					Error:            false,
					Success:          true,
				}, err
			}
		} else {
			logger.Log.Info("object is same on the server", zap.String("container", operatorContainer.Meta.Name))
		}

		return implementations.Response{
			HttpStatus:       200,
			Explanation:      "everything went smoothly: good job!",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
		}, nil
	}
}

func (implementation *Implementation) generateContainerNameAndGroup(mgr *manager.Manager, containerDefinition definitions.Container, changelog diff.Changelog) ([]string, []string, error) {
	_, index := mgr.Registry.Name(containerDefinition.Meta.Group, containerDefinition.Meta.Name, mgr.Runtime.PROJECT)

	r := replicas.Replicas{
		Group:          containerDefinition.Meta.Group,
		GeneratedIndex: index - 1,
		Replicas:       containerDefinition.Spec.Container.Replicas,
	}

	groups, names, err := r.HandleContainer(mgr, containerDefinition, changelog)

	return groups, names, err
}

// Exported
var Operator Implementation
