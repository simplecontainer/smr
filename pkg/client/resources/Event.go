package resources

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/contexts"
	"github.com/simplecontainer/smr/pkg/network"
	"net/http"
)

func Event(context *contexts.ClientContext, prefix string, version string, category string, kind string, group string, name string, data []byte) {
	response := network.Send(context.GetHTTPClient(), fmt.Sprintf("%s/api/v1/kind/propose/%s/%s/%s/%s/%s/%s", context.APIURL, prefix, version, category, kind, group, name), http.MethodPost, data)

	if response.Success {
		fmt.Println(response.Explanation)
	} else {
		fmt.Println(response.ErrorExplanation)
	}
}
