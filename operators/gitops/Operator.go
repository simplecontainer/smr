package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/qdnqn/smr/pkg/operators"
	"reflect"
)

func (operator *Operator) Run(operation string, args ...interface{}) operators.Response {
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

			return returnValue[0].Interface().(operators.Response)
		}
	}

	return operators.Response{
		HttpStatus:       400,
		Explanation:      "server doesn't support requested functionality",
		ErrorExplanation: "implementation is missing",
		Error:            true,
		Success:          false,
		Data:             nil,
	}
}

func (operator *Operator) ListSupported(args ...interface{}) operators.Response {
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

	return operators.Response{
		HttpStatus:       200,
		Explanation:      "",
		ErrorExplanation: "",
		Error:            true,
		Success:          false,
		Data:             supportedOperations,
	}
}

func (operator *Operator) List(request operators.Request) operators.Response {
	data := make(map[string]any)
	for key, gitops := range request.Manager.RepositoryWatchers.Repositories {
		data[key] = gitops
	}

	return operators.Response{
		HttpStatus:       200,
		Explanation:      "list of the gitops objects",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             data,
	}
}

func (operator *Operator) Sync(request operators.Request) operators.Response {
	return operators.Response{
		HttpStatus:       200,
		Explanation:      "sync is triggered",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             nil,
	}
}

// Exported
var Gitops Operator
