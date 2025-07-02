package resources

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/network"
	"net/http"
)

func listResources(context *client.ClientContext, endpoint string) ([]json.RawMessage, error) {
	response := network.Send(context.GetClient(), fmt.Sprintf("%s%s", context.APIURL, endpoint), http.MethodGet, nil)

	if response.HttpStatus != http.StatusOK || response.Error {
		return nil, errors.New(response.ErrorExplanation)
	}

	objects := make([]json.RawMessage, 0)
	err := json.Unmarshal(response.Data, &objects)
	if err != nil {
		return nil, err
	}

	return objects, nil
}

func ListKind(context *client.ClientContext, prefix string, version string, category string, kind string) ([]json.RawMessage, error) {
	endpoint := fmt.Sprintf("/api/v1/kind/%s/%s/%s/%s", prefix, version, category, kind)
	return listResources(context, endpoint)
}

func ListState(context *client.ClientContext, prefix string, version string, category string, kind string) ([]json.RawMessage, error) {
	endpoint := fmt.Sprintf("/api/v1/state/%s/%s/%s/%s", prefix, version, category, kind)
	return listResources(context, endpoint)
}

func ListKindGroup(context *client.ClientContext, prefix string, version string, category string, kind string, group string) ([]json.RawMessage, error) {
	endpoint := fmt.Sprintf("/api/v1/kind/%s/%s/%s/%s/%s", prefix, version, category, kind, group)
	return listResources(context, endpoint)
}
