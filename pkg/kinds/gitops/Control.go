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
	return common.Response(http.StatusOK, "", nil, network.ToJson(supportedControlOperations))
}

func (gitops *Gitops) List(request contracts.Control) contracts.Response {
	data := make(map[string]any)

	for key, gitopsInstance := range gitops.Shared.Watcher.Repositories {
		data[key] = gitopsInstance.Gitops
	}

	return common.Response(http.StatusOK, "", nil, network.ToJson(data))
}

func (gitops *Gitops) Get(request contracts.Control) contracts.Response {
	format := f.NewFromString(fmt.Sprintf("%s.%s.%s.%s", KIND, request.Group, request.Name, "object"))

	obj := objects.New(gitops.Shared.Client.Get(request.User.Username), request.User)
	obj.Find(format)

	if !obj.Exists() {
		return common.Response(http.StatusNotFound, static.STATUS_RESPONSE_NOT_FOUND, errors.New(static.STATUS_RESPONSE_NOT_FOUND), nil)
	}

	r, err := common.NewRequest(KIND)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil)
	}

	err = r.Definition.FromJson(obj.GetDefinitionByte())

	if err != nil {
		return contracts.Response{}
	}

	bytes, err := r.Definition.ToJsonWithKind()

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil)
	}

	return common.Response(http.StatusOK, "", nil, bytes)
}

func (gitops *Gitops) Remove(data contracts.Control) contracts.Response {
	request, err := common.NewRequest(static.KIND_GITOPS)

	if err != nil {
		return common.Response(http.StatusBadRequest, static.STATUS_RESPONSE_BAD_REQUEST, err, nil)
	}

	format := f.New("gitops", data.Group, data.Name, "object")
	obj := objects.New(gitops.Shared.Client.Get(data.User.Username), data.User)

	_, err = request.Definition.Delete(format, obj, static.KIND_GITOPS)

	if err != nil {
		return common.Response(http.StatusInternalServerError, static.STATUS_RESPONSE_INTERNAL_ERROR, err, nil)
	}

	GroupIdentifier := fmt.Sprintf("%s.%s", data.Group, data.Name)
	gitopsWatcher := gitops.Shared.Watcher.Find(GroupIdentifier)

	if gitopsWatcher == nil {
		return common.Response(http.StatusNotFound, "gitops definition doesn't exists", errors.New("gitops definition doesn't exists"), nil)
	} else {
		gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_PENDING_DELETE)
		gitopsWatcher.GitopsQueue <- gitopsWatcher.Gitops
	}

	return common.Response(http.StatusOK, static.STATUS_RESPONSE_DELETED, nil, nil)
}

func (gitops *Gitops) Refresh(request contracts.Control) contracts.Response {
	GroupIdentifier := fmt.Sprintf("%s.%s", request.Group, request.Name)
	gitopsWatcher := gitops.Shared.Watcher.Find(GroupIdentifier)

	if gitopsWatcher == nil {
		return common.Response(http.StatusNotFound, static.STATUS_RESPONSE_NOT_FOUND, errors.New("gitops doesn't exist"), nil)
	} else {
		gitopsWatcher.Gitops.ForcePoll = true
		gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_CLONING_GIT)
		gitopsWatcher.GitopsQueue <- gitopsWatcher.Gitops
	}

	return common.Response(http.StatusOK, static.STATUS_RESPONSE_REFRESHED, nil, nil)
}
func (gitops *Gitops) Sync(request contracts.Control) contracts.Response {
	GroupIdentifier := fmt.Sprintf("%s.%s", request.Group, request.Name)
	gitopsWatcher := gitops.Shared.Watcher.Find(GroupIdentifier)

	if gitopsWatcher == nil {
		return common.Response(http.StatusNotFound, static.STATUS_RESPONSE_NOT_FOUND, errors.New("gitops doesn't exist"), nil)
	} else {
		if gitopsWatcher.Gitops.AutomaticSync == false {
			gitopsWatcher.Gitops.ManualSync = true
		}

		gitopsWatcher.Gitops.ForcePoll = true
		gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_CLONING_GIT)
		gitopsWatcher.GitopsQueue <- gitopsWatcher.Gitops
	}

	return common.Response(http.StatusOK, static.STATUS_RESPONSE_SYNCED, nil, nil)
}
