package main

import (
	"database/sql"
	"fmt"
	"reflect"
	"smr/pkg/operators"
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
	for i := 0; i < reflected.NumMethod(); i++ {
		method := reflected.Method(i)

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

func (operator *Operator) DatabaseReady(request operators.Request) operators.Response {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/?timeout=5s", request.Data["username"], request.Data["password"], request.Data["ip"], request.Data["port"]))
	defer db.Close()

	if err != nil {
		return operators.Response{
			HttpStatus:       400,
			Explanation:      "database connection can't be opened",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	err = db.Ping()

	if err != nil {
		return operators.Response{
			HttpStatus:       400,
			Explanation:      "database can't be pinged",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	return operators.Response{
		HttpStatus:       200,
		Explanation:      "database connection is ready",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             nil,
	}
}

// Exported
var Mysql Operator
