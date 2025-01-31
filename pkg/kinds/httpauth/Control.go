package httpauth

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/static"
	"net/http"
)

var supportedControlOperations = []string{"List", "Get", "Remove"}

func (httpauth *Httpauth) ListSupported(request contracts.Control) contracts.Response {
	return common.Response(http.StatusOK, "", nil, network.ToJson(supportedControlOperations))
}
func (httpauth *Httpauth) List(request contracts.Control) contracts.Response {
	data := make(map[string]any)

	format := f.New(KIND, "", "", "")

	obj := objects.New(httpauth.Shared.Client.Get(request.User.Username), request.User)
	objs, err := obj.FindMany(format)

	if err != nil {
		return common.Response(http.StatusInternalServerError, static.STATUS_RESPONSE_INTERNAL_ERROR, err, nil)
	}

	for k, v := range objs {
		data[k] = v.GetDefinition()
	}

	return common.Response(http.StatusOK, "", nil, network.ToJson(data))
}
func (httpauth *Httpauth) Get(request contracts.Control) contracts.Response {
	format := f.NewFromString(fmt.Sprintf("/%s/%s/%s/%s", KIND, request.Group, request.Name, "object"))

	obj := objects.New(httpauth.Shared.Client.Get(request.User.Username), request.User)
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
func (httpauth *Httpauth) Remove(request contracts.Control) contracts.Response {
	GroupIdentifier := fmt.Sprintf("%s.%s", request.Group, request.Name)
	format := f.NewFromString(GroupIdentifier)

	obj := objects.New(httpauth.Shared.Client.Get(request.User.Username), request.User)
	err := obj.Find(format)

	if err != nil {
		return common.Response(http.StatusNotFound, static.STATUS_RESPONSE_NOT_FOUND, err, nil)
	}

	err = obj.Propose(format, nil)

	if err != nil {
		return common.Response(http.StatusInternalServerError, static.STATUS_RESPONSE_INTERNAL_ERROR, err, nil)
	} else {
		return common.Response(http.StatusOK, static.STATUS_RESPONSE_DELETED, nil, nil)
	}
}
