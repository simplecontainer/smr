package resources

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/contexts"
	"github.com/simplecontainer/smr/pkg/contracts/iresponse"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/static"
	"net/http"
	"strings"
)

func Inspect(context *contexts.ClientContext, prefix string, version string, category string, kind string, group string, field string) (json.RawMessage, error) {
	var response *iresponse.Response

	if kind == static.KIND_CONTAINERS {
		tmp := strings.Split(field, "-")
		name := strings.Join(tmp[1:len(tmp)-1], "-")

		response = network.Send(context.GetHTTPClient(), fmt.Sprintf("%s/api/v1/state/%s/%s/state/%s/%s/%s/%s", context.APIURL, prefix, version, kind, group, name, field), http.MethodGet, nil)
	} else {
		response = network.Send(context.GetHTTPClient(), fmt.Sprintf("%s/api/v1/state/%s/%s/state/%s/%s/%s/%s", context.APIURL, prefix, version, kind, group, field, field), http.MethodGet, nil)
	}

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
