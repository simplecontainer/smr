package resources

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/network"
	"net/http"
)

func Delete(context *client.ClientContext, prefix string, version string, category string, kind string, group string, name string) error {
	response := network.Send(context.GetClient(), fmt.Sprintf("%s/api/v1/kind/%s/%s/%s/%s/%s/%s", context.APIURL, prefix, version, category, kind, group, name), http.MethodGet, nil)

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

		return request.ProposeRemove(context.GetClient(), context.APIURL)
	} else {
		return errors.New(response.ErrorExplanation)
	}
}
