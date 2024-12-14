package resource

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/objects"
)

var supportedControlOperations = []string{"List", "Get", "Remove"}

func (resource *Resource) ListSupported(request contracts.Control) contracts.Response {
	return contracts.Response{
		HttpStatus:       200,
		Explanation:      "",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             network.ToJson(supportedControlOperations),
	}
}

func (resource *Resource) List(request contracts.Control) contracts.Response {
	data := make(map[string]any)

	format := f.New(KIND, "", "", "")

	obj := objects.New(resource.Shared.Client.Get(request.User.Username), request.User)
	objs, err := obj.FindMany(format)

	if err != nil {
		return contracts.Response{
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

	return contracts.Response{
		HttpStatus:       200,
		Explanation:      "list of the resource objects",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             network.ToJson(data),
	}
}

func (resource *Resource) Get(request contracts.Control) contracts.Response {
	format := f.NewFromString(fmt.Sprintf("%s.%s.%s.%s", KIND, request.Group, request.Group, "object"))

	obj := objects.New(resource.Shared.Client.Get(request.User.Username), request.User)
	err := obj.Find(format)

	if err != nil {
		return contracts.Response{
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

	return contracts.Response{
		HttpStatus:       200,
		Explanation:      "gitops object is found on the server",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             network.ToJson(definition),
	}
}

func (resource *Resource) Remove(request contracts.Control) contracts.Response {
	GroupIdentifier := fmt.Sprintf("%s.%s", request.Group, request.Name)
	format := f.NewFromString(GroupIdentifier)

	obj := objects.New(resource.Shared.Client.Get(request.User.Username), request.User)
	err := obj.Find(format)

	if err != nil {
		return contracts.Response{
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
		return contracts.Response{
			HttpStatus:       500,
			Explanation:      "resource definition is not deleted",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	} else {
		return contracts.Response{
			HttpStatus:       200,
			Explanation:      "resource definition is deleted and removed from server",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
			Data:             nil,
		}
	}
}
