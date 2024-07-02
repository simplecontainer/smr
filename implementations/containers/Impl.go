package main

import "C"
import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/qdnqn/smr/implementations/containers/reconcile"
	"github.com/qdnqn/smr/implementations/containers/shared"
	"github.com/qdnqn/smr/pkg/database"
	"github.com/qdnqn/smr/pkg/definitions/v1"
	"github.com/qdnqn/smr/pkg/httpcontract"
	"github.com/qdnqn/smr/pkg/manager"
	"github.com/qdnqn/smr/pkg/objects"
)

func (implementation *Implementation) Start(mgr *manager.Manager) error {
	implementation.Shared.Manager = mgr
	return nil
}

func (implementation *Implementation) Apply(jsonData []byte) (httpcontract.ResponseImplementation, error) {
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
	err = obj.Find(implementation.Shared.Manager.Badger, format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = containersDefinition.ToJsonString()

	if obj.Exists() {
		if obj.Diff(jsonStringFromRequest) {
			err = obj.Update(implementation.Shared.Manager.Badger, format, jsonStringFromRequest)
		}
	} else {
		err = obj.Add(implementation.Shared.Manager.Badger, format, jsonStringFromRequest)
	}

	if obj.ChangeDetected() || !obj.Exists() {
		containersFromDefinition := reconcile.NewWatcher(*containersDefinition)
		GroupIdentifier := fmt.Sprintf("%s.%s", containersDefinition.Meta.Group, containersDefinition.Meta.Name)

		implementation.Shared.Watcher.AddOrUpdate(GroupIdentifier, containersFromDefinition)
		go reconcile.HandleTickerAndEvents(implementation.Shared, containersFromDefinition)

		reconcile.ReconcileContainer(implementation.Shared, containersFromDefinition)
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
	err = obj.Find(implementation.Shared.Manager.Badger, format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = containersDefinition.ToJsonString()

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
	err = obj.Find(implementation.Shared.Manager.Badger, format)

	if !obj.Exists() {
		return httpcontract.ResponseImplementation{
			HttpStatus:       404,
			Explanation:      "object not found on the server",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
		}, nil
	}

	obj.Remove(implementation.Shared.Manager.Badger, format)

	GroupIdentifier := fmt.Sprintf("%s.%s", containersDefinition.Meta.Group, containersDefinition.Meta.Name)
	implementation.Shared.Watcher.Find(GroupIdentifier).Cancel()
	implementation.Shared.Watcher.Remove(GroupIdentifier)

	for _, definition := range containersDefinition.Spec {
		format := database.Format("container", definition.Meta.Group, definition.Meta.Name, "object")
		obj := objects.New()
		err = obj.Find(implementation.Shared.Manager.Badger, format)

		if err != nil {
		}
		/*
			if obj.Exists() {
				groups, names, _ := reconcile.GetReplicaNamesAndGroups(shared, definition, obj.Changelog)

				if len(groups) > 0 {
					containers := reconcile.FetchContainersFromRegistry(implementation.Shared.Registry, groups, names)

					for _, containerObj := range containers {
						containerObj.Status.TransitionState(status.STATUS_PENDING_DELETE)

						format = database.Format("runtime", containerObj.Static.Group, containerObj.Static.GeneratedName, "")
						obj.Remove(implementation.Shared.Manager.Badger, format)

						format = database.Format("configuration", containerObj.Static.Group, containerObj.Static.GeneratedName, "")
						obj.Remove(implementation.Shared.Manager.Badger, format)
					}
				}

				obj.Remove(implementation.Shared.Manager.Badger, format)

				format := database.Format("container", definition.Meta.Group, definition.Meta.Name, "")
				obj.Remove(implementation.Shared.Manager.Badger, format)
			}

		*/
	}

	return httpcontract.ResponseImplementation{
		HttpStatus:       200,
		Explanation:      "everything went good",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
	}, nil
}

// Exported
var Containers Implementation = Implementation{
	Started: false,
	Shared:  &shared.Shared{},
}
