package certkey

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/httpcontract"
	"github.com/simplecontainer/smr/pkg/objects"
	"reflect"
)

func (certkey *Certkey) Run(operation string, args ...interface{}) httpcontract.ResponseOperator {
	reflected := reflect.TypeOf(certkey)
	reflectedValue := reflect.ValueOf(certkey)

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

func (certkey *Certkey) ListSupported(args ...interface{}) httpcontract.ResponseOperator {
	reflected := reflect.TypeOf(certkey)

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

func (certkey *Certkey) List(request httpcontract.RequestOperator) httpcontract.ResponseOperator {
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

func (certkey *Certkey) Get(request httpcontract.RequestOperator) httpcontract.ResponseOperator {
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
			Explanation:      "certkey definition is not found on the server",
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
		Explanation:      "certkey object is found on the server",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             definition,
	}
}

func (certkey *Certkey) Delete(request httpcontract.RequestOperator) httpcontract.ResponseOperator {
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

	GroupIdentifier := fmt.Sprintf("%s.%s", request.Data["group"], request.Data["identifier"])
	format := f.NewFromString(GroupIdentifier)

	obj := objects.New(request.Client.Get(request.User.Username), request.User)
	err := obj.Find(format)

	if err != nil {
		return httpcontract.ResponseOperator{
			HttpStatus:       404,
			Explanation:      "certkey definition is not found on the server",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	removed, err := obj.Remove(format)

	if !removed {
		return httpcontract.ResponseOperator{
			HttpStatus:       500,
			Explanation:      "certkey definition is not deleted",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	} else {
		return httpcontract.ResponseOperator{
			HttpStatus:       200,
			Explanation:      "certkey definition is deleted and removed from server",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
			Data:             nil,
		}
	}
}
