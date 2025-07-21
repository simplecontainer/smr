package resources

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/contexts"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/network"
	"net/http"
)

func Edit(context *contexts.ClientContext, prefix string, version string, category string, kind string, group string, name string) (json.RawMessage, error) {
	response := network.Send(context.GetHTTPClient(), fmt.Sprintf("%s/api/v1/kind/%s/%s/%s/%s/%s/%s", context.APIURL, prefix, version, category, kind, group, name), http.MethodGet, nil)

	object := json.RawMessage{}

	err := json.Unmarshal(response.Data, &object)

	if err != nil {
		return nil, err
	}

	data, changed, err := helpers.Editor(object)

	if err != nil {
		return nil, err
	}

	request, err := common.NewRequest(kind)

	if err != nil {
		return nil, err
	}

	err = request.Definition.FromJson(data)

	if err != nil {
		return nil, err
	}

	if changed {
		err = request.ProposeApply(context.GetHTTPClient(), context.APIURL)
		return data, err
	}

	return nil, errors.New("nothing changed")
}
