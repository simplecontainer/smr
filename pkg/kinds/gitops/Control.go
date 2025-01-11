package gitops

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/status"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/static"
	"net/http"
)

var supportedControlOperations = []string{"List", "Get", "Remove", "Refresh", "Sync"}

func (gitops *Gitops) ListSupported(request contracts.Control) contracts.Response {
	return contracts.Response{
		HttpStatus:       200,
		Explanation:      "",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             network.ToJson(supportedControlOperations),
	}
}
func (gitops *Gitops) List(request contracts.Control) contracts.Response {
	data := make(map[string]any)

	for key, gitopsInstance := range gitops.Shared.Watcher.Repositories {
		data[key] = gitopsInstance.Gitops
	}

	return contracts.Response{
		HttpStatus:       200,
		Explanation:      "list of the gitops objects",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             network.ToJson(data),
	}
}
func (gitops *Gitops) Get(request contracts.Control) contracts.Response {
	format := f.NewFromString(fmt.Sprintf("%s.%s.%s.%s", KIND, request.Group, request.Name, "object"))

	obj := objects.New(gitops.Shared.Client.Get(request.User.Username), request.User)
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

func (gitops *Gitops) Remove(data contracts.Control) contracts.Response {
	request, err := common.NewRequest(static.KIND_GITOPS)

	if err != nil {
		return common.Response(http.StatusBadRequest, err.Error(), err)
	}

	format := f.New("gitops", data.Group, data.Name, "object")
	obj := objects.New(gitops.Shared.Client.Get(data.User.Username), data.User)

	_, err = request.Definition.Delete(format, obj, static.KIND_GITOPS)

	if err != nil {
		return common.Response(http.StatusInternalServerError, err.Error(), err)
	}

	GroupIdentifier := fmt.Sprintf("%s.%s", data.Group, data.Name)
	gitopsWatcher := gitops.Shared.Watcher.Find(GroupIdentifier)

	if gitopsWatcher == nil {
		return common.Response(http.StatusNotFound, "gitops definition doesn't exists", errors.New("gitops definition doesn't exists"))
	} else {
		gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_PENDING_DELETE)
		gitopsWatcher.GitopsQueue <- gitopsWatcher.Gitops
	}

	return common.Response(http.StatusOK, "object deleted", nil)
}

func (gitops *Gitops) Refresh(request contracts.Control) contracts.Response {
	GroupIdentifier := fmt.Sprintf("%s.%s", request.Group, request.Name)
	gitopsWatcher := gitops.Shared.Watcher.Find(GroupIdentifier)

	if gitopsWatcher == nil {
		return contracts.Response{
			HttpStatus:       404,
			Explanation:      "gitops definition doesn't exists",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	} else {
		gitopsWatcher.Gitops.ForcePoll = true
		gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_CLONING_GIT)
		gitopsWatcher.GitopsQueue <- gitopsWatcher.Gitops
	}

	return contracts.Response{
		HttpStatus:       200,
		Explanation:      "refresh is triggered manually",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             nil,
	}
}
func (gitops *Gitops) Sync(request contracts.Control) contracts.Response {
	GroupIdentifier := fmt.Sprintf("%s.%s", request.Group, request.Name)
	gitopsWatcher := gitops.Shared.Watcher.Find(GroupIdentifier)

	if gitopsWatcher == nil {
		return contracts.Response{
			HttpStatus:       404,
			Explanation:      "gitops definition doesn't exists",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	} else {
		if gitopsWatcher.Gitops.AutomaticSync == false {
			gitopsWatcher.Gitops.ManualSync = true
		}

		gitopsWatcher.Gitops.ForcePoll = true
		gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_CLONING_GIT)
		gitopsWatcher.GitopsQueue <- gitopsWatcher.Gitops
	}

	return contracts.Response{
		HttpStatus:       200,
		Explanation:      "sync is triggered manually",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             nil,
	}
}
