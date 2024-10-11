package main

import "C"
import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/implementations/containers/reconcile"
	"github.com/simplecontainer/smr/implementations/containers/shared"
	"github.com/simplecontainer/smr/implementations/containers/watcher"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/httpcontract"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/plugins"
	"go.uber.org/zap"
	"net/http"
)

func (implementation *Implementation) Start(mgr *manager.Manager) error {
	implementation.Shared.Manager = mgr
	implementation.Started = true

	implementation.Shared.Client = mgr.Http

	implementation.Shared.Watcher = &watcher.ContainersWatcher{}
	implementation.Shared.Watcher.Containers = make(map[string]*watcher.Containers)

	return nil
}

func (implementation *Implementation) GetShared() interface{} {
	return implementation.Shared
}

func (implementation *Implementation) Apply(user *authentication.User, jsonData []byte) (httpcontract.ResponseImplementation, error) {
	containersDefinition := &v1.ContainersDefinition{}

	if err := json.Unmarshal(jsonData, &containersDefinition); err != nil {
		return httpcontract.ResponseImplementation{
			HttpStatus:       400,
			Explanation:      "invalid definition sent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}, err
	}

	valid, err := containersDefinition.Validate()

	if !valid {
		return httpcontract.ResponseImplementation{
			HttpStatus:       400,
			Explanation:      "invalid definition sent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}, err
	}

	data := make(map[string]interface{})
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		panic(err)
	}

	var format *f.Format
	format = f.New("containers", containersDefinition.Meta.Group, containersDefinition.Meta.Name, "object")

	obj := objects.New(implementation.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = containersDefinition.ToJsonString()

	logger.Log.Debug("server received containers object", zap.String("definition", jsonStringFromRequest))

	if obj.Exists() {
		if obj.Diff(jsonStringFromRequest) {
			err = obj.Update(format, jsonStringFromRequest)

			if err != nil {
				return httpcontract.ResponseImplementation{
					HttpStatus:       http.StatusInternalServerError,
					Explanation:      "failed to update object",
					ErrorExplanation: err.Error(),
					Error:            false,
					Success:          true,
				}, err
			}
		}
	} else {
		err = obj.Add(format, jsonStringFromRequest)

		if err != nil {
			return httpcontract.ResponseImplementation{
				HttpStatus:       http.StatusInternalServerError,
				Explanation:      "failed to add object",
				ErrorExplanation: err.Error(),
				Error:            false,
				Success:          true,
			}, err
		}
	}

	GroupIdentifier := fmt.Sprintf("%s.%s", containersDefinition.Meta.Group, containersDefinition.Meta.Name)
	containersFromDefinition := implementation.Shared.Watcher.Find(GroupIdentifier)

	if obj.Exists() {
		if obj.ChangeDetected() || containersFromDefinition == nil {
			if containersFromDefinition == nil {
				containersFromDefinition = reconcile.NewWatcher(*containersDefinition, implementation.Shared.Manager)
				containersFromDefinition.Logger.Info("containers object created")

				go reconcile.HandleTickerAndEvents(implementation.Shared, user, containersFromDefinition)
			} else {
				containersFromDefinition.Definition = *containersDefinition
				containersFromDefinition.Logger.Info("containers object modified")
			}

			implementation.Shared.Watcher.AddOrUpdate(GroupIdentifier, containersFromDefinition)
			reconcile.ReconcileContainer(implementation.Shared, user, containersFromDefinition)
		} else {
			return httpcontract.ResponseImplementation{
				HttpStatus:       http.StatusOK,
				Explanation:      "containers object is same as the one on the server",
				ErrorExplanation: "",
				Error:            false,
				Success:          true,
			}, errors.New("containers object is same on the server")
		}
	} else {
		containersFromDefinition = reconcile.NewWatcher(*containersDefinition, implementation.Shared.Manager)
		containersFromDefinition.Logger.Info("containers object created")

		go reconcile.HandleTickerAndEvents(implementation.Shared, user, containersFromDefinition)

		implementation.Shared.Watcher.AddOrUpdate(GroupIdentifier, containersFromDefinition)
		reconcile.ReconcileContainer(implementation.Shared, user, containersFromDefinition)
	}

	return httpcontract.ResponseImplementation{
		HttpStatus:       http.StatusOK,
		Explanation:      "everything went smoothly: good job!",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
	}, nil
}

func (implementation *Implementation) Compare(user *authentication.User, jsonData []byte) (httpcontract.ResponseImplementation, error) {
	containersDefinition := &v1.ContainersDefinition{}

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

	var format *f.Format
	format = f.New("containers", containersDefinition.Meta.Group, containersDefinition.Meta.Name, "object")

	obj := objects.New(implementation.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

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

func (implementation *Implementation) Delete(user *authentication.User, jsonData []byte) (httpcontract.ResponseImplementation, error) {
	containersDefinition := &v1.ContainersDefinition{}

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

	var format *f.Format
	format = f.New("containers", containersDefinition.Meta.Group, containersDefinition.Meta.Name, "object")

	obj := objects.New(implementation.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	if !obj.Exists() {
		return httpcontract.ResponseImplementation{
			HttpStatus:       404,
			Explanation:      "",
			ErrorExplanation: "object not found on the server",
			Error:            true,
			Success:          false,
		}, nil
	}

	_, err = obj.Remove(format)

	if err != nil {
		return httpcontract.ResponseImplementation{
			HttpStatus:       500,
			Explanation:      "object removal failed",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}, nil
	}

	GroupIdentifier := fmt.Sprintf("%s.%s", containersDefinition.Meta.Group, containersDefinition.Meta.Name)

	implementation.Shared.Watcher.Find(GroupIdentifier).Syncing = true
	implementation.Shared.Watcher.Find(GroupIdentifier).Cancel()

	for _, definition := range containersDefinition.Spec {
		format = f.New("container", definition.Meta.Group, definition.Meta.Name, "object")
		obj = objects.New(implementation.Shared.Client.Get(user.Username), user)
		obj.Find(format)

		if obj.Exists() {
			pl := plugins.GetPlugin(implementation.Shared.Manager.Config.OptRoot, "container.so")
			pl.Delete(user, obj.GetDefinitionByte())
		}
	}

	return httpcontract.ResponseImplementation{
		HttpStatus:       200,
		Explanation:      "action completed successfully",
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
