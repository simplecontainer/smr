package containers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/container/shared"
	"github.com/simplecontainer/smr/pkg/kinds/containers/reconcile"
	"github.com/simplecontainer/smr/pkg/kinds/containers/watcher"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/objects"
	"go.uber.org/zap"
	"net/http"
	"reflect"
)

func (containers *Containers) Start() error {
	containers.Started = true

	containers.Shared.Watcher = &watcher.ContainersWatcher{}
	containers.Shared.Watcher.Containers = make(map[string]*watcher.Containers)

	return nil
}

func (containers *Containers) GetShared() interface{} {
	return containers.Shared
}
func (containers *Containers) Apply(user *authentication.User, jsonData []byte) (contracts.ResponseImplementation, error) {
	containersDefinition := &v1.ContainersDefinition{}

	if err := json.Unmarshal(jsonData, &containersDefinition); err != nil {
		return contracts.ResponseImplementation{
			HttpStatus:       400,
			Explanation:      "invalid definition sent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}, err
	}

	valid, err := containersDefinition.Validate()

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

	var format *f.Format
	format = f.New("containers", containersDefinition.Meta.Group, containersDefinition.Meta.Name, "object")

	obj := objects.New(containers.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = containersDefinition.ToJsonString()

	logger.Log.Debug("server received containers object", zap.String("definition", jsonStringFromRequest))

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

	GroupIdentifier := fmt.Sprintf("%s.%s", containersDefinition.Meta.Group, containersDefinition.Meta.Name)
	containersFromDefinition := containers.Shared.Watcher.Find(GroupIdentifier)

	if obj.Exists() {
		if obj.ChangeDetected() || containersFromDefinition == nil {
			if containersFromDefinition == nil {
				containersFromDefinition = reconcile.NewWatcher(*containersDefinition, containers.Shared.Manager)
				containersFromDefinition.Logger.Info("containers object created")

				go reconcile.HandleTickerAndEvents(containers.Shared, user, containersFromDefinition)
			} else {
				containersFromDefinition.Definition = *containersDefinition
				containersFromDefinition.Logger.Info("containers object modified")
			}

			containers.Shared.Watcher.AddOrUpdate(GroupIdentifier, containersFromDefinition)
			reconcile.Container(containers.Shared, user, containersFromDefinition)
		} else {
			return contracts.ResponseImplementation{
				HttpStatus:       http.StatusOK,
				Explanation:      "containers object is same as the one on the server",
				ErrorExplanation: "",
				Error:            false,
				Success:          true,
			}, errors.New("containers object is same on the server")
		}
	} else {
		containersFromDefinition = reconcile.NewWatcher(*containersDefinition, containers.Shared.Manager)
		containersFromDefinition.Logger.Info("containers object created")

		go reconcile.HandleTickerAndEvents(containers.Shared, user, containersFromDefinition)

		containers.Shared.Watcher.AddOrUpdate(GroupIdentifier, containersFromDefinition)
		reconcile.Container(containers.Shared, user, containersFromDefinition)
	}

	return contracts.ResponseImplementation{
		HttpStatus:       http.StatusOK,
		Explanation:      "everything went smoothly: good job!",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
	}, nil
}
func (containers *Containers) Compare(user *authentication.User, jsonData []byte) (contracts.ResponseImplementation, error) {
	containersDefinition := &v1.ContainersDefinition{}

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
	format = f.New("containers", containersDefinition.Meta.Group, containersDefinition.Meta.Name, "object")

	obj := objects.New(containers.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = containersDefinition.ToJsonString()

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
func (containers *Containers) Delete(user *authentication.User, jsonData []byte) (contracts.ResponseImplementation, error) {
	containersDefinition := &v1.ContainersDefinition{}

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
	format = f.New("containers", containersDefinition.Meta.Group, containersDefinition.Meta.Name, "object")

	obj := objects.New(containers.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	if !obj.Exists() {
		return contracts.ResponseImplementation{
			HttpStatus:       404,
			Explanation:      "",
			ErrorExplanation: "object not found on the server",
			Error:            true,
			Success:          false,
		}, nil
	}

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

	GroupIdentifier := fmt.Sprintf("%s.%s", containersDefinition.Meta.Group, containersDefinition.Meta.Name)

	if containers.Shared.Watcher.Find(GroupIdentifier) != nil {
		containers.Shared.Watcher.Find(GroupIdentifier).Syncing = true
		containers.Shared.Watcher.Find(GroupIdentifier).Cancel()
	}

	for _, definition := range containersDefinition.Spec {
		format = f.New("container", definition.Meta.Group, definition.Meta.Name, "object")
		obj = objects.New(containers.Shared.Client.Get(user.Username), user)
		obj.Find(format)

		if obj.Exists() {
			// 50 cent
			//pl := plugins.GetPlugin(containers.Shared.Manager.Config.OptRoot, "container.so")
			//pl.Delete(user, obj.GetDefinitionByte())
		}
	}

	return contracts.ResponseImplementation{
		HttpStatus:       200,
		Explanation:      "action completed successfully",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
	}, nil
}

func (containers *Containers) Run(operation string, args ...interface{}) contracts.ResponseOperator {
	reflected := reflect.TypeOf(containers)
	reflectedValue := reflect.ValueOf(containers)

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
func (containers *Containers) ListSupported(args ...interface{}) contracts.ResponseOperator {
	reflected := reflect.TypeOf(containers)

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
func (containers *Containers) List(request contracts.RequestOperator) contracts.ResponseOperator {
	data := make(map[string]any)

	format := f.New(KIND, "", "", "")

	obj := objects.New(containers.Shared.Client.Get(request.User.Username), request.User)
	objs, err := obj.FindMany(format)

	if err != nil {
		return contracts.ResponseOperator{
			HttpStatus:       400,
			Explanation:      "error occured",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	for k, v := range objs {
		data[k] = v.GetDefinition()
	}

	return contracts.ResponseOperator{
		HttpStatus:       200,
		Explanation:      "list of the certkey objects",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             data,
	}
}
func (containers *Containers) Get(request contracts.RequestOperator) contracts.ResponseOperator {
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

	obj := objects.New(containers.Shared.Client.Get(request.User.Username), request.User)
	err := obj.Find(format)

	if err != nil {
		return contracts.ResponseOperator{
			HttpStatus:       404,
			Explanation:      "container definition is not found on the server",
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
		Explanation:      "container object is found on the server",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             definition,
	}
}
func (containers *Containers) View(request contracts.RequestOperator) contracts.ResponseOperator {
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

	registry := containers.Shared.Manager.KindsRegistry["container"].GetShared().(shared.Shared)
	container := registry.Registry.Find(fmt.Sprintf("%s", request.Data["group"]), fmt.Sprintf("%s", request.Data["identifier"]))

	if container == nil {
		return contracts.ResponseOperator{
			HttpStatus:       404,
			Explanation:      "container not found in the registry",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	var definition = make(map[string]any)
	definition[container.GetGeneratedName()] = container

	return contracts.ResponseOperator{
		HttpStatus:       200,
		Explanation:      "container object is found on the server",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             nil,
	}
}
