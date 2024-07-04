package main

import "C"
import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/implementations/containers/reconcile"
	"github.com/simplecontainer/smr/implementations/containers/shared"
	"github.com/simplecontainer/smr/implementations/containers/watcher"
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

	implementation.Shared.Client = client

	implementation.Shared.Watcher = &watcher.ContainersWatcher{}
	implementation.Shared.Watcher.Containers = make(map[string]*watcher.Containers)

	return nil
}

func (implementation *Implementation) GetShared() interface{} {
	return implementation.Shared
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

	var format objects.FormatStructure
	format = objects.Format("containers", containersDefinition.Meta.Group, containersDefinition.Meta.Name, "object")

	obj := objects.New()
	err = obj.Find(implementation.Shared.Client, format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = containersDefinition.ToJsonString()

	logger.Log.Debug("server received containers object", zap.String("definition", jsonStringFromRequest))

	if obj.Exists() {
		if obj.Diff(jsonStringFromRequest) {
			err = obj.Update(implementation.Shared.Client, format, jsonStringFromRequest)
		}
	} else {
		err = obj.Add(implementation.Shared.Client, format, jsonStringFromRequest)
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

	var format objects.FormatStructure
	format = objects.Format("containers", containersDefinition.Meta.Group, containersDefinition.Meta.Name, "object")

	obj := objects.New()
	err = obj.Find(implementation.Shared.Client, format)

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

	var format objects.FormatStructure
	format = objects.Format("containers", containersDefinition.Meta.Group, containersDefinition.Meta.Name, "object")

	obj := objects.New()
	err = obj.Find(implementation.Shared.Client, format)

	if !obj.Exists() {
		return httpcontract.ResponseImplementation{
			HttpStatus:       404,
			Explanation:      "object not found on the server",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
		}, nil
	}

	obj.Remove(implementation.Shared.Client, format)

	GroupIdentifier := fmt.Sprintf("%s.%s", containersDefinition.Meta.Group, containersDefinition.Meta.Name)
	implementation.Shared.Watcher.Find(GroupIdentifier).Cancel()
	implementation.Shared.Watcher.Remove(GroupIdentifier)

	for _, definition := range containersDefinition.Spec {
		format = objects.Format("container", definition.Meta.Group, definition.Meta.Name, "object")
		obj = objects.New()
		obj.Find(implementation.Shared.Client, format)

		if obj.Exists() {
			pl := plugins.GetPlugin(implementation.Shared.Manager.Config.Root, "container.so")
			pl.Delete(obj.GetDefinitionByte())
		}
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
