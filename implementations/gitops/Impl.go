package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/simplecontainer/smr/implementations/gitops/gitops"
	"github.com/simplecontainer/smr/implementations/gitops/reconcile"
	"github.com/simplecontainer/smr/implementations/gitops/shared"
	"github.com/simplecontainer/smr/implementations/gitops/status"
	"github.com/simplecontainer/smr/implementations/gitops/watcher"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/httpcontract"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/objects"
	"go.uber.org/zap"
	"net/http"
)

func (implementation *Implementation) Start(mgr *manager.Manager) error {
	implementation.Shared.Manager = mgr
	implementation.Started = true

	implementation.Shared.Client = mgr.Http

	implementation.Shared.Watcher = &watcher.RepositoryWatcher{}
	implementation.Shared.Watcher.Repositories = make(map[string]*watcher.Gitops)

	return nil
}

func (implementation *Implementation) GetShared() interface{} {
	return implementation.Shared
}

func (implementation *Implementation) Apply(user *authentication.User, jsonData []byte) (httpcontract.ResponseImplementation, error) {
	var gitopsDefinition = &v1.GitopsDefinition{}

	if err := json.Unmarshal(jsonData, &gitopsDefinition); err != nil {
		return httpcontract.ResponseImplementation{
			HttpStatus:       400,
			Explanation:      "invalid configuration sent: json is not valid",
			ErrorExplanation: "invalid configuration sent: json is not valid",
			Error:            true,
			Success:          false,
		}, err
	}

	valid, err := gitopsDefinition.Validate()

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

	mapstructure.Decode(data, &gitopsDefinition)

	var format *f.Format

	format = f.New("gitops", gitopsDefinition.Meta.Group, gitopsDefinition.Meta.Name, "object")
	obj := objects.New(implementation.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = gitopsDefinition.ToJsonString()

	logger.Log.Debug("server received gitops object", zap.String("definition", jsonStringFromRequest))

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

	GroupIdentifier := fmt.Sprintf("%s.%s", gitopsDefinition.Meta.Group, gitopsDefinition.Meta.Name)
	gitopsWatcherFromRegistry := implementation.Shared.Watcher.Find(GroupIdentifier)

	if obj.Exists() {
		if obj.ChangeDetected() || gitopsWatcherFromRegistry == nil {
			if gitopsWatcherFromRegistry == nil {
				gitopsWatcherFromRegistry = reconcile.NewWatcher(gitops.New(gitopsDefinition), implementation.Shared.Manager, user)
				go reconcile.HandleTickerAndEvents(implementation.Shared, gitopsWatcherFromRegistry)

				gitopsWatcherFromRegistry.Logger.Info("new gitops object created")
				gitopsWatcherFromRegistry.Gitops.Status.SetState(status.STATUS_CREATED)
			} else {
				implementation.Shared.Watcher.Find(GroupIdentifier).Gitops = gitops.New(gitopsDefinition)
				gitopsWatcherFromRegistry.Logger.Info("gitops object modified")
				gitopsWatcherFromRegistry.Gitops.Status.SetState(status.STATUS_CREATED)
			}

			implementation.Shared.Watcher.AddOrUpdate(GroupIdentifier, gitopsWatcherFromRegistry)
			reconcile.Gitops(implementation.Shared, gitopsWatcherFromRegistry)
		} else {
			return httpcontract.ResponseImplementation{
				HttpStatus:       http.StatusOK,
				Explanation:      "object is same on the server",
				ErrorExplanation: "",
				Error:            false,
				Success:          true,
			}, errors.New("object is same on the server")
		}
	} else {
		gitopsWatcherFromRegistry = reconcile.NewWatcher(gitops.New(gitopsDefinition), implementation.Shared.Manager, user)
		go reconcile.HandleTickerAndEvents(implementation.Shared, gitopsWatcherFromRegistry)

		gitopsWatcherFromRegistry.Logger.Info("new gitops object created")
		gitopsWatcherFromRegistry.Gitops.Status.SetState(status.STATUS_CREATED)
		implementation.Shared.Watcher.AddOrUpdate(GroupIdentifier, gitopsWatcherFromRegistry)
		reconcile.Gitops(implementation.Shared, gitopsWatcherFromRegistry)
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
	var gitopsDefinition v1.GitopsDefinition

	if err := json.Unmarshal(jsonData, &gitopsDefinition); err != nil {
		return httpcontract.ResponseImplementation{
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

	mapstructure.Decode(data["gitops"], &gitopsDefinition)

	var format *f.Format

	format = f.New("gitops", gitopsDefinition.Meta.Group, gitopsDefinition.Meta.Name, "object")
	obj := objects.New(implementation.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = gitopsDefinition.ToJsonString()

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
	containersDefinition := &v1.GitopsDefinition{}

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
	format = f.New("gitops", containersDefinition.Meta.Group, containersDefinition.Meta.Name, "object")

	obj := objects.New(implementation.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	if !obj.Exists() {
		return httpcontract.ResponseImplementation{
			HttpStatus:       404,
			Explanation:      "object not found on the server",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
		}, nil
	}

	GroupIdentifier := fmt.Sprintf("%s.%s", containersDefinition.Meta.Group, containersDefinition.Meta.Name)

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

	gitopsObj := implementation.Shared.Watcher.Find(GroupIdentifier).Gitops

	gitopsObj.Status.TransitionState(gitopsObj.Definition.Meta.Name, status.STATUS_PENDING_DELETE)
	reconcile.Gitops(implementation.Shared, implementation.Shared.Watcher.Find(GroupIdentifier))

	return httpcontract.ResponseImplementation{
		HttpStatus:       200,
		Explanation:      "object in sync",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
	}, nil
}

var Gitops Implementation = Implementation{
	Started: false,
	Shared:  &shared.Shared{},
}
