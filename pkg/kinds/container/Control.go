package container

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/container/status"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/objects"
	"net/http"
)

var supportedControlOperations = []string{"List", "Get", "Remove", "View", "Restart"}

func (container *Container) ListSupported(request contracts.Control) contracts.Response {
	return contracts.Response{
		HttpStatus:       200,
		Explanation:      "",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             network.ToJson(supportedControlOperations),
	}
}

func (container *Container) List(request contracts.Control) contracts.Response {
	data := make(map[string]any)

	format := f.New(KIND, "", "", "")

	obj := objects.New(container.Shared.Client.Get(request.User.Username), request.User)
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
func (container *Container) Get(request contracts.Control) contracts.Response {
	format := f.NewFromString(fmt.Sprintf("%s.%s.%s.%s", KIND, request.Group, request.Name, "object"))

	obj := objects.New(container.Shared.Client.Get(request.User.Username), request.User)
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
		Data:             network.ToJson(definition),
	}
}
func (container *Container) View(request contracts.Control) contracts.Response {
	containerObj := container.Shared.Registry.FindLocal(fmt.Sprintf("%s", request.Group), fmt.Sprintf("%s", request.Name))

	if containerObj == nil {
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
	definition[containerObj.GetGeneratedName()] = containerObj

	return contracts.Response{
		HttpStatus:       200,
		Explanation:      "container object is found on the server",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             network.ToJson(definition),
	}
}
func (container *Container) Restart(request contracts.Control) contracts.Response {
	containerObj := container.Shared.Registry.FindLocal(fmt.Sprintf("%s", request.Group), fmt.Sprintf("%s", request.Name))

	if containerObj == nil {
		return contracts.Response{
			HttpStatus:       404,
			Explanation:      "container not found in the registry",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_CREATED)
	container.Shared.Watcher.Find(containerObj.GetGroupIdentifier()).ContainerQueue <- containerObj

	return contracts.Response{
		HttpStatus:       200,
		Explanation:      "container object is restarted",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             nil,
	}
}
func (container *Container) Remove(request contracts.Control) contracts.Response {
	format := f.New("container", request.Group, request.Name, "object")

	obj := objects.New(container.Shared.Client.Get(request.User.Username), request.User)
	err := obj.Find(format)
	if err != nil {
		panic(err)
	}

	if !obj.Exists() {
		return contracts.Response{
			HttpStatus:       404,
			Explanation:      "object not found on the server",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
		}
	}

	_, err = container.Delete(request.User, obj.GetDefinitionByte(), container.Shared.Manager.Config.Node)

	if err != nil {
		return contracts.Response{
			HttpStatus:       http.StatusInternalServerError,
			Explanation:      "failed to delete containers",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}
	}

	return contracts.Response{
		HttpStatus:       http.StatusOK,
		Explanation:      "action completed successfully",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
	}
}
