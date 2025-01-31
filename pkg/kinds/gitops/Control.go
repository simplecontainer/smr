package gitops

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/KV"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/events"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/implementation"
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
	var reg map[string]map[string]*implementation.Gitops
	reg = gitops.Shared.Registry.All()

	if len(reg) > 0 {
		return common.Response(http.StatusOK, "", nil, network.ToJson(reg))
	} else {
		return common.Response(http.StatusOK, "", nil, nil)
	}
}

func (gitops *Gitops) Get(request contracts.Control) contracts.Response {
	format, _ := f.NewFromString(fmt.Sprintf("/%s/%s/%s/%s", KIND, request.Group, request.Name, "object"))

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

	bytes, err := r.Definition.ToJsonForUser()

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

	format, _ := f.New("gitops", data.Group, data.Name, "object")
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
	gitopsObj := gitops.Shared.Registry.Find(request.Group, request.Name)

	if gitopsObj != nil {

		event := events.New(events.EVENT_REFRESH, static.KIND_GITOPS, gitopsObj.GetGroup(), gitopsObj.GetName(), nil)

		bytes, err := event.ToJson()

		if err != nil {
			return common.Response(http.StatusInternalServerError, static.STATUS_RESPONSE_INTERNAL_ERROR, err, nil)
		}

		gitops.Shared.Manager.Cluster.KVStore.Propose(event.GetKey(), bytes, gitopsObj.Definition.GetRuntime().GetNode())
		return common.Response(http.StatusOK, static.STATUS_RESPONSE_REFRESHED, nil, nil)
	} else {
		return common.Response(http.StatusNotFound, static.STATUS_RESPONSE_NOT_FOUND, nil, nil)
	}
}

func (gitops *Gitops) Sync(request contracts.Control) contracts.Response {
	gitopsObj := gitops.Shared.Registry.Find(request.Group, request.Name)

	if gitopsObj != nil {
		event := events.New(events.EVENT_SYNC, static.KIND_GITOPS, gitopsObj.GetGroup(), gitopsObj.GetName(), nil)

		bytes, err := event.ToJson()

		if err != nil {
			return common.Response(http.StatusInternalServerError, static.STATUS_RESPONSE_INTERNAL_ERROR, err, nil)
		}

		gitops.Shared.Manager.Replication.EventsC <- KV.NewEncode(event.GetKey(), bytes, gitops.Shared.Manager.Config.KVStore.Node)
		gitops.Shared.Manager.Cluster.KVStore.Propose(event.GetKey(), bytes, gitopsObj.Definition.GetRuntime().GetNode())

		return common.Response(http.StatusOK, static.STATUS_RESPONSE_SYNCED, nil, nil)
	} else {
		return common.Response(http.StatusNotFound, static.STATUS_RESPONSE_NOT_FOUND, nil, nil)
	}
}
