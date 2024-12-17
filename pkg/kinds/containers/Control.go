package containers

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/container/shared"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/objects"
)

var supportedControlOperations = []string{"List", "Get", "View"}

func (containers *Containers) ListSupported(request contracts.Control) contracts.Response {
	return contracts.Response{
		HttpStatus:       200,
		Explanation:      "",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             network.ToJson(supportedControlOperations),
	}
}

func (containers *Containers) List(request contracts.Control) contracts.Response {
	data := make(map[string]any)

	format := f.New(KIND, "", "", "")

	obj := objects.New(containers.Shared.Client.Get(request.User.Username), request.User)
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
		Explanation:      "list of the certkey objects",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             network.ToJson(data),
	}
}
func (containers *Containers) Get(request contracts.Control) contracts.Response {
	format := f.NewFromString(fmt.Sprintf("%s.%s.%s.%s", KIND, request.Group, request.Name, "object"))

	obj := objects.New(containers.Shared.Client.Get(request.User.Username), request.User)
	err := obj.Find(format)

	if err != nil {
		return contracts.Response{
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

	return contracts.Response{
		HttpStatus:       200,
		Explanation:      "container object is found on the server",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             network.ToJson(network.ToJson(definition)),
	}
}
func (containers *Containers) View(request contracts.Control) contracts.Response {
	registry := containers.Shared.Manager.KindsRegistry["container"].GetShared().(shared.Shared)
	container := registry.Registry.Find(fmt.Sprintf("%s", request.Group), fmt.Sprintf("%s", request.Name))

	if container == nil {
		return contracts.Response{
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

	return contracts.Response{
		HttpStatus:       200,
		Explanation:      "container object is found on the server",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             nil,
	}
}
