package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
	"github.com/qdnqn/smr/pkg/database"
	"github.com/qdnqn/smr/pkg/definitions/v1"
	"github.com/qdnqn/smr/pkg/gitops"
	"github.com/qdnqn/smr/pkg/httpcontract"
	"github.com/qdnqn/smr/pkg/logger"
	"github.com/qdnqn/smr/pkg/manager"
	"github.com/qdnqn/smr/pkg/objects"
)

func (implementation *Implementation) Apply(mgr *manager.Manager, jsonData []byte, c *gin.Context) (httpcontract.ResponseImplementation, error) {
	var gitopsDefinition v1.Gitops

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

	var format database.FormatStructure

	format = database.Format("gitops", gitopsDefinition.Meta.Group, gitopsDefinition.Meta.Identifier, "object")
	obj := objects.New()
	err = obj.Find(mgr.Registry.Object, mgr.Badger, format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = gitopsDefinition.ToJsonString()

	if obj.Exists() {
		if obj.Diff(jsonStringFromRequest) {
			err = obj.Update(mgr.Registry.Object, mgr.Badger, format, jsonStringFromRequest)
		}
	} else {
		err = obj.Add(mgr.Registry.Object, mgr.Badger, format, jsonStringFromRequest)
	}

	if obj.ChangeDetected() || !obj.Exists() {
		gitopsFromDefinition := gitops.NewWatcher(gitopsDefinition)
		GroupIdentifier := fmt.Sprintf("%s.%s", gitopsFromDefinition.Definition.Meta.Group, gitopsFromDefinition.Definition.Meta.Identifier)

		mgr.RepositoryWatchers.AddOrUpdate(GroupIdentifier, gitopsFromDefinition)
		go mgr.RepositoryWatchers.Find(GroupIdentifier).HandleTickerAndEvents(mgr.DefinitionRegistry, mgr.Keys)
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
	var gitopsDefinition v1.Gitops

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

	var format database.FormatStructure

	format = database.Format("gitops", gitopsDefinition.Meta.Group, gitopsDefinition.Meta.Identifier, "object")
	obj := objects.New()
	err = obj.Find(mgr.Registry.Object, mgr.Badger, format)

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

func (implementation *Implementation) Delete(mgr *manager.Manager, jsonData []byte, c *gin.Context) (httpcontract.ResponseImplementation, error) {
	var gitopsDefinition v1.Gitops

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

	format := database.Format("gitops", gitopsDefinition.Meta.Group, gitopsDefinition.Meta.Identifier, "object")

	obj := objects.New()
	err = obj.Find(mgr.Registry.Object, mgr.Badger, format)

	if obj.Exists() {
		logger.Log.Info("removing gitops reconcile")

		GroupIdentifier := fmt.Sprintf("%s.%s", gitopsDefinition.Meta.Group, gitopsDefinition.Meta.Identifier)
		mgr.RepositoryWatchers.Find(GroupIdentifier).GitopsQueue <- gitops.Event{
			Event: gitops.KILL,
		}

		if mgr.RepositoryWatchers.Remove(GroupIdentifier) {
			deleted, err := obj.Remove(mgr.Badger, format)

			if deleted {
				return httpcontract.ResponseImplementation{
					HttpStatus:       200,
					Explanation:      "deleted resource successfully",
					ErrorExplanation: "",
					Error:            true,
					Success:          false,
				}, err
			} else {
				return httpcontract.ResponseImplementation{
					HttpStatus:       500,
					Explanation:      "failed to delete resource",
					ErrorExplanation: err.Error(),
					Error:            true,
					Success:          false,
				}, err
			}
		} else {
			return httpcontract.ResponseImplementation{
				HttpStatus:       500,
				Explanation:      "failed to stop gitops reconcile",
				ErrorExplanation: "",
				Error:            true,
				Success:          false,
			}, nil
		}
	} else {
		return httpcontract.ResponseImplementation{
			HttpStatus:       404,
			Explanation:      "object not found",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
		}, err
	}
}

var Gitops Implementation
