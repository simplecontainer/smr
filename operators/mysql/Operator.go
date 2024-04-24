package main

import (
	"database/sql"
	"fmt"
	"reflect"
)

func (operator *Operator) Run(operation string, args ...interface{}) map[string]any {
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

			return returnValue[0].Interface().(map[string]any)
		}
	}

	return map[string]any{
		"message": "Operator doesn't have that functionality.",
	}
}

func (operator *Operator) ListSupported(args ...interface{}) map[string]any {
	reflected := reflect.TypeOf(operator)

	supportedOperations := map[string]any{}
	supportedOperations["SupportedOperations"] = []string{}
	for i := 0; i < reflected.NumMethod(); i++ {
		method := reflected.Method(i)

		supportedOperations["SupportedOperations"] = append(supportedOperations["SupportedOperations"].([]string), method.Name)
	}

	return supportedOperations
}

func (operator *Operator) DatabaseReady(data map[string]any) map[string]any {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/?timeout=5s", data["username"], data["password"], data["ip"], data["port"]))
	defer db.Close()

	if err != nil {
		return map[string]any{
			"message": err.Error(),
			"success": "false",
		}
	}

	err = db.Ping()

	if err != nil {
		return map[string]any{
			"message": err.Error(),
			"success": "false",
		}
	}

	return map[string]any{
		"success": "true",
	}
}

// Exported
var Mysql Operator
