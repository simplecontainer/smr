package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/simplecontainer/smr/implementations/container/shared"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/httpcontract"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/operators"
	"github.com/simplecontainer/smr/pkg/plugins"
	"reflect"
)

func (operator *Operator) Run(operation string, args ...interface{}) httpcontract.ResponseOperator {
	reflected := reflect.TypeOf(operator)
	reflectedValue := reflect.ValueOf(operator)

	for i := 0; i < reflected.NumMethod(); i++ {
		method := reflected.Method(i)

		if operation == method.Name {
			inputs := make([]reflect.Value, len(args))

			for i, _ := range args {
				inputs[i] = reflect.ValueOf(args[i])
			}

			returnValue := reflectedValue.MethodByName(operation).Call(inputs)

			return returnValue[0].Interface().(httpcontract.ResponseOperator)
		}
	}

	return httpcontract.ResponseOperator{
		HttpStatus:       400,
		Explanation:      "server doesn't support requested functionality",
		ErrorExplanation: "implementation is missing",
		Error:            true,
		Success:          false,
		Data:             nil,
	}
}

func (operator *Operator) ListSupported(args ...interface{}) httpcontract.ResponseOperator {
	reflected := reflect.TypeOf(operator)

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

	return httpcontract.ResponseOperator{
		HttpStatus:       200,
		Explanation:      "",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             supportedOperations,
	}
}

func (operator *Operator) List(request operators.Request) httpcontract.ResponseOperator {
	data := make(map[string]any)

	format := f.New(KIND, "", "", "")

	obj := objects.New(request.Client.Get(request.User.Username), request.User)
	objs, err := obj.FindMany(format)

	if err != nil {
		return httpcontract.ResponseOperator{
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

	return httpcontract.ResponseOperator{
		HttpStatus:       200,
		Explanation:      "list of the certkey objects",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             data,
	}
}

func (operator *Operator) Get(request operators.Request) httpcontract.ResponseOperator {
	if request.Data == nil {
		return httpcontract.ResponseOperator{
			HttpStatus:       400,
			Explanation:      "send some data",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	format := f.NewFromString(fmt.Sprintf("%s.%s.%s.%s", KIND, request.Data["group"], request.Data["identifier"], "object"))

	obj := objects.New(request.Client.Get(request.User.Username), request.User)
	err := obj.Find(format)

	if err != nil {
		return httpcontract.ResponseOperator{
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

	return httpcontract.ResponseOperator{
		HttpStatus:       200,
		Explanation:      "container object is found on the server",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             definition,
	}
}

func (operator *Operator) View(request operators.Request) httpcontract.ResponseOperator {
	if request.Data == nil {
		return httpcontract.ResponseOperator{
			HttpStatus:       400,
			Explanation:      "send some data",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	pl := plugins.GetPlugin(request.Manager.Config.OptRoot, "container.so")
	sharedObj := pl.GetShared().(*shared.Shared)

	container := sharedObj.Registry.Find(fmt.Sprintf("%s", request.Data["group"]), fmt.Sprintf("%s", request.Data["identifier"]))

	if container == nil {
		return httpcontract.ResponseOperator{
			HttpStatus:       404,
			Explanation:      "container not found in the registry",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	var definition = make(map[string]any)
	definition[container.Static.GeneratedName] = container

	return httpcontract.ResponseOperator{
		HttpStatus:       200,
		Explanation:      "container object is found on the server",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             definition,
	}
}

// Exported
var Containers Operator
