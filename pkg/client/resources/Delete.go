package resources

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/simplecontainer/smr/pkg/contexts"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/network"
	"net/http"
)

func Delete(context *contexts.ClientContext, prefix string, version string, category string, kind string, group string, name string) error {
	response := network.Send(context.GetHTTPClient(), fmt.Sprintf("%s/api/v1/kind/%s/%s/%s/%s/%s/%s", context.APIURL, prefix, version, category, kind, group, name), http.MethodGet, nil)

	object := json.RawMessage{}

	err := json.Unmarshal(response.Data, &object)

	if err != nil {
		return err
	}

	if response.HttpStatus == http.StatusOK {
		request, err := common.NewRequest(kind)

		if err != nil {
			return err
		}

		err = request.Definition.FromJson(response.Data)

		if err != nil {
			return err
		}

		return request.ProposeRemove(context.GetHTTPClient(), context.APIURL)
	} else {
		return errors.New(response.ErrorExplanation)
	}
}
