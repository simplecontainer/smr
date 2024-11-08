package gitops

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/implementation"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/reconcile"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/status"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/watcher"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/objects"
	"go.uber.org/zap"
	"net/http"
	"reflect"
)

func (gitops *Gitops) Start() error {
	gitops.Started = true

	gitops.Shared.Watcher = &watcher.RepositoryWatcher{
		Repositories: make(map[string]*watcher.Gitops),
	}

	return nil
}
func (gitops *Gitops) GetShared() interface{} {
	return gitops.Shared
}
func (gitops *Gitops) Apply(user *authentication.User, jsonData []byte) (contracts.ResponseImplementation, error) {
	var gitopsDefinition = &v1.GitopsDefinition{}

	if err := json.Unmarshal(jsonData, &gitopsDefinition); err != nil {
		return contracts.ResponseImplementation{
			HttpStatus:       400,
			Explanation:      "invalid configuration sent: json is not valid",
			ErrorExplanation: "invalid configuration sent: json is not valid",
			Error:            true,
			Success:          false,
		}, err
	}

	valid, err := gitopsDefinition.Validate()

	if !valid {
		return contracts.ResponseImplementation{
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
	obj := objects.New(gitops.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = gitopsDefinition.ToJsonString()

	logger.Log.Debug("server received gitops object", zap.String("definition", jsonStringFromRequest))

	if obj.Exists() {
		if obj.Diff(jsonStringFromRequest) {
			err = obj.Update(format, jsonStringFromRequest)

			if err != nil {
				return contracts.ResponseImplementation{
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
			return contracts.ResponseImplementation{
				HttpStatus:       http.StatusInternalServerError,
				Explanation:      "failed to add object",
				ErrorExplanation: err.Error(),
				Error:            false,
				Success:          true,
			}, err
		}
	}

	GroupIdentifier := fmt.Sprintf("%s.%s", gitopsDefinition.Meta.Group, gitopsDefinition.Meta.Name)
	gitopsWatcherFromRegistry := gitops.Shared.Watcher.Find(GroupIdentifier)

	if obj.Exists() {
		if obj.ChangeDetected() || gitopsWatcherFromRegistry == nil {
			if gitopsWatcherFromRegistry == nil {
				gitopsWatcherFromRegistry = reconcile.NewWatcher(implementation.New(gitopsDefinition), gitops.Shared.Manager, user)
				go reconcile.HandleTickerAndEvents(gitops.Shared, gitopsWatcherFromRegistry)

				gitopsWatcherFromRegistry.Logger.Info("new gitops object created")
				gitopsWatcherFromRegistry.Gitops.Status.SetState(status.STATUS_CREATED)
			} else {
				gitops.Shared.Watcher.Find(GroupIdentifier).Gitops = implementation.New(gitopsDefinition)
				gitopsWatcherFromRegistry.Logger.Info("gitops object modified")
				gitopsWatcherFromRegistry.Gitops.Status.SetState(status.STATUS_CREATED)
			}

			gitops.Shared.Watcher.AddOrUpdate(GroupIdentifier, gitopsWatcherFromRegistry)
			reconcile.Gitops(gitops.Shared, gitopsWatcherFromRegistry)
		} else {
			return contracts.ResponseImplementation{
				HttpStatus:       http.StatusOK,
				Explanation:      "object is same on the server",
				ErrorExplanation: "",
				Error:            false,
				Success:          true,
			}, errors.New("object is same on the server")
		}
	} else {
		gitopsWatcherFromRegistry = reconcile.NewWatcher(implementation.New(gitopsDefinition), gitops.Shared.Manager, user)
		go reconcile.HandleTickerAndEvents(gitops.Shared, gitopsWatcherFromRegistry)

		gitopsWatcherFromRegistry.Logger.Info("new gitops object created")
		gitopsWatcherFromRegistry.Gitops.Status.SetState(status.STATUS_CREATED)
		gitops.Shared.Watcher.AddOrUpdate(GroupIdentifier, gitopsWatcherFromRegistry)
		reconcile.Gitops(gitops.Shared, gitopsWatcherFromRegistry)
	}

	return contracts.ResponseImplementation{
		HttpStatus:       http.StatusOK,
		Explanation:      "everything went smoothly: good job!",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
	}, nil
}
func (gitops *Gitops) Compare(user *authentication.User, jsonData []byte) (contracts.ResponseImplementation, error) {
	var gitopsDefinition v1.GitopsDefinition

	if err := json.Unmarshal(jsonData, &gitopsDefinition); err != nil {
		return contracts.ResponseImplementation{
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
	obj := objects.New(gitops.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = gitopsDefinition.ToJsonString()

	if obj.Exists() {
		obj.Diff(jsonStringFromRequest)
	} else {
		return contracts.ResponseImplementation{
			HttpStatus:       418,
			Explanation:      "object is drifted from the definition",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
		}, nil
	}

	if obj.ChangeDetected() {
		return contracts.ResponseImplementation{
			HttpStatus:       418,
			Explanation:      "object is drifted from the definition",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
		}, nil
	} else {
		return contracts.ResponseImplementation{
			HttpStatus:       200,
			Explanation:      "object in sync",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
		}, nil
	}
}
func (gitops *Gitops) Delete(user *authentication.User, jsonData []byte) (contracts.ResponseImplementation, error) {
	containersDefinition := &v1.GitopsDefinition{}

	if err := json.Unmarshal(jsonData, &containersDefinition); err != nil {
		return contracts.ResponseImplementation{
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

	obj := objects.New(gitops.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	if !obj.Exists() {
		return contracts.ResponseImplementation{
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
		return contracts.ResponseImplementation{
			HttpStatus:       500,
			Explanation:      "object removal failed",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}, nil
	}

	gitopsObj := gitops.Shared.Watcher.Find(GroupIdentifier).Gitops

	gitopsObj.Status.TransitionState(gitopsObj.Definition.Meta.Name, status.STATUS_PENDING_DELETE)
	reconcile.Gitops(gitops.Shared, gitops.Shared.Watcher.Find(GroupIdentifier))

	return contracts.ResponseImplementation{
		HttpStatus:       200,
		Explanation:      "object in sync",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
	}, nil
}
func (gitops *Gitops) Run(operation string, args ...interface{}) contracts.ResponseOperator {
	reflected := reflect.TypeOf(gitops)
	reflectedValue := reflect.ValueOf(gitops)

	for i := 0; i < reflected.NumMethod(); i++ {
		method := reflected.Method(i)

		if operation == method.Name {
			inputs := make([]reflect.Value, len(args))

			for i, _ := range args {
				inputs[i] = reflect.ValueOf(args[i])
			}

			returnValue := reflectedValue.MethodByName(operation).Call(inputs)

			return returnValue[0].Interface().(contracts.ResponseOperator)
		}
	}

	return contracts.ResponseOperator{
		HttpStatus:       400,
		Explanation:      "server doesn't support requested functionality",
		ErrorExplanation: "implementation is missing",
		Error:            true,
		Success:          false,
		Data:             nil,
	}
}

func (gitops *Gitops) ListSupported(args ...interface{}) contracts.ResponseOperator {
	reflected := reflect.TypeOf(gitops)

	supportedOperations := map[string]any{}
	supportedOperations["SupportedOperations"] = []string{}

OUTER:
	for i := 0; i < reflected.NumMethod(); i++ {
		method := reflected.Method(i)
		for _, forbiddenOperator := range invalidOperators {
			if forbiddenOperator == method.Name {
				continue OUTER
			}
		}

		supportedOperations["SupportedOperations"] = append(supportedOperations["SupportedOperations"].([]string), method.Name)
	}

	return contracts.ResponseOperator{
		HttpStatus:       200,
		Explanation:      "",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             supportedOperations,
	}
}
func (gitops *Gitops) List(request contracts.RequestOperator) contracts.ResponseOperator {
	data := make(map[string]any)

	for key, gitopsInstance := range gitops.Shared.Watcher.Repositories {
		data[key] = gitopsInstance.Gitops
	}

	return contracts.ResponseOperator{
		HttpStatus:       200,
		Explanation:      "list of the gitops objects",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             data,
	}
}
func (gitops *Gitops) Get(request contracts.RequestOperator) contracts.ResponseOperator {
	if request.Data == nil {
		return contracts.ResponseOperator{
			HttpStatus:       400,
			Explanation:      "send some data",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	format := f.NewFromString(fmt.Sprintf("%s.%s.%s.%s", KIND, request.Data["group"], request.Data["identifier"], "object"))

	obj := objects.New(gitops.Shared.Client.Get(request.User.Username), request.User)
	err := obj.Find(format)

	if err != nil {
		return contracts.ResponseOperator{
			HttpStatus:       404,
			Explanation:      "gitops definition is not found on the server",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	definitionObject := obj.GetDefinition()

	var definition = make(map[string]any)
	definition["kind"] = KIND
	definition[KIND] = definitionObject

	return contracts.ResponseOperator{
		HttpStatus:       200,
		Explanation:      "gitops object is found on the server",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             definition,
	}
}
func (gitops *Gitops) Remove(request contracts.RequestOperator) contracts.ResponseOperator {
	if request.Data == nil {
		return contracts.ResponseOperator{
			HttpStatus:       400,
			Explanation:      "send some data",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	format := f.New("gitops", request.Data["group"].(string), request.Data["identifier"].(string), "object")

	obj := objects.New(gitops.Shared.Client.Get(request.User.Username), request.User)
	err := obj.Find(format)

	if err != nil {
		return contracts.ResponseOperator{
			HttpStatus:       http.StatusInternalServerError,
			Explanation:      "object database failed to process request",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}
	}

	if !obj.Exists() {
		return contracts.ResponseOperator{
			HttpStatus:       404,
			Explanation:      "object not found on the server",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
		}
	}

	_, err = obj.Remove(format)

	if err != nil {
		return contracts.ResponseOperator{
			HttpStatus:       http.StatusInternalServerError,
			Explanation:      "object removal failed",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}
	}

	GroupIdentifier := fmt.Sprintf("%s.%s", request.Data["group"], request.Data["identifier"])
	gitopsWatcher := gitops.Shared.Watcher.Find(GroupIdentifier)

	if gitopsWatcher == nil {
		return contracts.ResponseOperator{
			HttpStatus:       404,
			Explanation:      "gitops definition doesn't exists",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	} else {
		gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_PENDING_DELETE)
		gitopsWatcher.GitopsQueue <- gitopsWatcher.Gitops
	}

	return contracts.ResponseOperator{
		HttpStatus:       200,
		Explanation:      "gitops is transitioned to the pending delete state",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             nil,
	}
}
func (gitops *Gitops) Refresh(request contracts.RequestOperator) contracts.ResponseOperator {
	if request.Data == nil {
		return contracts.ResponseOperator{
			HttpStatus:       400,
			Explanation:      "send some data",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	GroupIdentifier := fmt.Sprintf("%s.%s", request.Data["group"], request.Data["identifier"])
	gitopsWatcher := gitops.Shared.Watcher.Find(GroupIdentifier)

	if gitopsWatcher == nil {
		return contracts.ResponseOperator{
			HttpStatus:       404,
			Explanation:      "gitops definition doesn't exists",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	} else {
		gitopsWatcher.Gitops.ForcePoll = true
		gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_CLONING_GIT)
		gitopsWatcher.GitopsQueue <- gitopsWatcher.Gitops
	}

	return contracts.ResponseOperator{
		HttpStatus:       200,
		Explanation:      "refresh is triggered manually",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             nil,
	}
}
func (gitops *Gitops) Sync(request contracts.RequestOperator) contracts.ResponseOperator {
	if request.Data == nil {
		return contracts.ResponseOperator{
			HttpStatus:       400,
			Explanation:      "send some data",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	GroupIdentifier := fmt.Sprintf("%s.%s", request.Data["group"], request.Data["identifier"])
	gitopsWatcher := gitops.Shared.Watcher.Find(GroupIdentifier)

	if gitopsWatcher == nil {
		return contracts.ResponseOperator{
			HttpStatus:       404,
			Explanation:      "gitops definition doesn't exists",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	} else {
		if gitopsWatcher.Gitops.AutomaticSync == false {
			gitopsWatcher.Gitops.ManualSync = true
		}

		gitopsWatcher.Gitops.ForcePoll = true
		gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_CLONING_GIT)
		gitopsWatcher.GitopsQueue <- gitopsWatcher.Gitops
	}

	return contracts.ResponseOperator{
		HttpStatus:       200,
		Explanation:      "sync is triggered manually",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             nil,
	}
}
