package resources

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/network"
	"net/http"
)

func ListKind(context *client.ClientContext, prefix string, version string, category string, kind string) ([]json.RawMessage, error) {
	response := network.Send(context.GetClient(), fmt.Sprintf("%s/api/v1/kind/%s/%s/%s/%s", context.APIURL, prefix, version, category, kind), http.MethodGet, nil)

	if response.HttpStatus != http.StatusOK || response.Error {
		return nil, errors.New(response.ErrorExplanation)
	} else {
		objects := make([]json.RawMessage, 0)

		err := json.Unmarshal(response.Data, &objects)

		if err != nil {
			return nil, err
		}

		return objects, nil
	}
}

func ListKindGroup(context *client.ClientContext, prefix string, version string, category string, kind string, group string) ([]json.RawMessage, error) {
	response := network.Send(context.GetClient(), fmt.Sprintf("%s/api/v1/kind/%s/%s/%s/%s/%s", context.APIURL, prefix, version, category, kind, group), http.MethodGet, nil)

	if response.HttpStatus != http.StatusOK {
		return nil, errors.New(response.ErrorExplanation)
	} else {
		objects := make([]json.RawMessage, 0)

		err := json.Unmarshal(response.Data, &objects)

		if err != nil {
			return nil, err
		}

		return objects, nil
	}
}
