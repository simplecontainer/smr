package network

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mitchellh/mapstructure"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/network/implementation"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/objects"
	"go.uber.org/zap"
	"net/http"
	"reflect"
)

func (network *Network) Start() error {
	network.Started = true
	return nil
}

func (network *Network) GetShared() interface{} {
	return network.Shared
}

func (network *Network) Apply(user *authentication.User, jsonData []byte) (contracts.ResponseImplementation, error) {
	var networkDefinition v1.NetworkDefinition

	if err := json.Unmarshal(jsonData, &networkDefinition); err != nil {
		return contracts.ResponseImplementation{
			HttpStatus:       400,
			Explanation:      "invalid configuration sent: json is not valid",
			ErrorExplanation: "invalid configuration sent: json is not valid",
			Error:            true,
			Success:          false,
		}, err
	}

	valid, err := networkDefinition.Validate()

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

	mapstructure.Decode(data["network"], &networkDefinition)

	var format *f.Format
	format = f.New("network", networkDefinition.Meta.Group, networkDefinition.Meta.Name, "object")

	obj := objects.New(network.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = networkDefinition.ToJsonString()

	logger.Log.Debug("server received network object", zap.String("definition", jsonStringFromRequest))

	if obj.Exists() {
		if obj.Diff(jsonStringFromRequest) {
			err = obj.Update(format, jsonStringFromRequest)

			if err != nil {
				return contracts.ResponseImplementation{
					HttpStatus:       http.StatusInternalServerError,
					Explanation:      "",
					ErrorExplanation: err.Error(),
					Error:            true,
					Success:          false,
				}, err
			}
		}
	} else {
		err = obj.Add(format, jsonStringFromRequest)

		if err != nil {
			return contracts.ResponseImplementation{
				HttpStatus:       http.StatusInternalServerError,
				Explanation:      "",
				ErrorExplanation: err.Error(),
				Error:            true,
				Success:          false,
			}, err
		}
	}

	var networkObj *implementation.Network

	if obj.ChangeDetected() || !obj.Exists() {
		networkObj = implementation.New(jsonData)
	} else {
		networkObj = implementation.New(obj.GetDefinitionByte())
	}

	err = networkObj.Create()

	if err != nil {
		return contracts.ResponseImplementation{
			HttpStatus:       http.StatusInternalServerError,
			Explanation:      "",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}, err
	}

	return contracts.ResponseImplementation{
		HttpStatus:       200,
		Explanation:      "everything went smoothly: good job!",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
	}, nil
}

func (network *Network) Compare(user *authentication.User, jsonData []byte) (contracts.ResponseImplementation, error) {
	if err := json.Unmarshal(jsonData, &network.Definition); err != nil {
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

	mapstructure.Decode(data["spec"], &network)

	var format *f.Format

	format = f.New("network", network.Definition.Meta.Group, network.Definition.Meta.Name, "object")
	obj := objects.New(network.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = network.Definition.ToJsonString()

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

func (network *Network) Delete(user *authentication.User, jsonData []byte) (contracts.ResponseImplementation, error) {
	if err := json.Unmarshal(jsonData, &network.Definition); err != nil {
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

	mapstructure.Decode(data["network"], &network)

	format := f.New("network", network.Definition.Meta.Group, network.Definition.Meta.Name, "object")

	obj := objects.New(network.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	if obj.Exists() {
		deleted, err := obj.Remove(format)

		if deleted {
			format = f.New("network", network.Definition.Meta.Group, network.Definition.Meta.Name, "")
			deleted, err = obj.Remove(format)

			return contracts.ResponseImplementation{
				HttpStatus:       200,
				Explanation:      "deleted resource successfully",
				ErrorExplanation: "",
				Error:            true,
				Success:          false,
			}, err
		} else {
			return contracts.ResponseImplementation{
				HttpStatus:       500,
				Explanation:      "failed to delete resource",
				ErrorExplanation: err.Error(),
				Error:            true,
				Success:          false,
			}, err
		}
	} else {
		return contracts.ResponseImplementation{
			HttpStatus:       404,
			Explanation:      "object not found",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
		}, err
	}
}

func (network *Network) Run(operation string, args ...interface{}) contracts.ResponseOperator {
	reflected := reflect.TypeOf(network)
	reflectedValue := reflect.ValueOf(network)

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

func (network *Network) ListSupported(args ...interface{}) contracts.ResponseOperator {
	reflected := reflect.TypeOf(network)

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

func (network *Network) List(request contracts.RequestOperator) contracts.ResponseOperator {
	data := make(map[string]any)

	format := f.New(KIND, "", "", "")

	obj := objects.New(network.Shared.Client.Get(request.User.Username), request.User)
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
		Explanation:      "list of the resource objects",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             data,
	}
}

func (network *Network) Get(request contracts.RequestOperator) contracts.ResponseOperator {
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

	obj := objects.New(network.Shared.Client.Get(request.User.Username), request.User)
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

func (network *Network) Remove(request contracts.RequestOperator) contracts.ResponseOperator {
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
	format := f.NewFromString(GroupIdentifier)

	obj := objects.New(network.Shared.Client.Get(request.User.Username), request.User)
	err := obj.Find(format)

	if err != nil {
		return contracts.ResponseOperator{
			HttpStatus:       404,
			Explanation:      "resource definition is not found on the server",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	removed, err := obj.Remove(format)

	if !removed {
		return contracts.ResponseOperator{
			HttpStatus:       500,
			Explanation:      "resource definition is not deleted",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	} else {
		return contracts.ResponseOperator{
			HttpStatus:       200,
			Explanation:      "resource definition is deleted and removed from server",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
			Data:             nil,
		}
	}
}
