package resources

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/contexts"
	"github.com/simplecontainer/smr/pkg/network"
	"net/http"
	"strings"
)

func Inspect(context *contexts.ClientContext, prefix string, version string, category string, kind string, group string, field string) (json.RawMessage, error) {
	tmp := strings.Split(field, "-")
	name := strings.Join(tmp[1:len(tmp)-1], "-")

	response := network.Send(context.GetHTTPClient(), fmt.Sprintf("%s/api/v1/kind/%s/%s/state/%s/%s/%s/%s", context.APIURL, prefix, version, kind, group, name, field), http.MethodGet, nil)

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
