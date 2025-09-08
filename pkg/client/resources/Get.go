package resources

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/contexts"
	"github.com/simplecontainer/smr/pkg/network"
	"net/http"
)

func Get(context *contexts.ClientContext, storePrefix string, prefix string, version string, category string, kind string, group string, name string) (json.RawMessage, error) {
	response := network.Send(context.GetHTTPClient(), fmt.Sprintf("%s/api/v1/%s/%s/%s/%s/%s/%s/%s", context.APIURL, storePrefix, prefix, version, category, kind, group, name), http.MethodGet, nil)

	if response.HttpStatus != http.StatusOK {
		return nil, errors.New(response.ErrorExplanation)
	} else {
		object := json.RawMessage{}

		err := json.Unmarshal(response.Data, &object)

		if err != nil {
			return nil, err
		}

		return object, nil
	}
}
