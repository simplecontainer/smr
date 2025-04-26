package api

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/simplecontainer/smr/pkg/stream"
	"net/http"
	"strconv"
	"strings"
)

func (api *Api) Debug(c *gin.Context) {
	prefix := c.Param("prefix")
	version := c.Param("version")
	category := c.Param("category")
	kind := c.Param("kind")
	group := c.Param("group")
	name := c.Param("name")
	which := c.Param("which")

	follow, err := strconv.ParseBool(c.Param("follow"))

	if err != nil {
		follow = false
	}

	format := f.New(prefix, version, category, kind, group, name)

	w := c.Writer
	header := w.Header()
	header.Set("Transfer-Encoding", "chunked")
	header.Set("Content-Type", "application/json")
	header.Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	if kind == static.KIND_CONTAINERS {
		container := api.KindsRegistry[static.KIND_CONTAINERS].GetShared().(*shared.Shared).Registry.Find(static.SMR_PREFIX, group, name)

		if container == nil {
			stream.Bye(w, errors.New("container is not found"))
			return
		} else {
			if container.IsGhost() {
				client, ok := api.Manager.Http.Clients[container.GetRuntime().Node.NodeName]
				if !ok {
					stream.Bye(w, errors.New(fmt.Sprintf("node for %s '%s/%s' not found", static.KIND_CONTAINER, group, name)))
					return
				}

				URL := fmt.Sprintf("%s/api/v1/debug/%s/%s/%s", client.API, format.ToString(), which, c.Param("follow"))
				err = stream.Remote(c, URL, client)

				if err != nil {
					stream.Bye(w, err)
				}

				return
			} else {
				err = stream.Tail(c, w, fmt.Sprintf("/tmp/%s", strings.Replace(format.ToString(), "/", "-", -1)), follow)

				if err != nil {
					stream.Bye(w, err)
					return
				}
			}
		}
	} else {
		format = f.New(prefix, version, "kind", kind, group, name)

		obj := objects.New(api.Manager.Http.Clients[api.User.Username], api.User)
		obj.Find(format)

		if !obj.Exists() {
			stream.Bye(w, errors.New(fmt.Sprintf("%s not found", kind)))
			return
		}

		var request *common.Request
		request, err = common.NewRequest(kind)

		if err != nil {
			stream.Bye(w, err)
			return
		}

		err = request.Definition.FromJson(obj.GetDefinitionByte())

		if request.Definition.GetRuntime().GetNode() != api.Cluster.Node.NodeID {
			client, ok := api.Manager.Http.Clients[request.Definition.GetRuntime().GetNodeName()]

			if !ok {
				stream.Bye(w, errors.New(fmt.Sprintf("%s is not found", kind)))
				return
			} else {
				URL := fmt.Sprintf("%s/api/v1/debug/%s/%s/%s", client.API, format.ToString(), which, c.Param("follow"))
				err = stream.Remote(c, URL, client)

				if err != nil {
					stream.Bye(w, err)
				}
			}
		} else {
			err = stream.Tail(c, w, fmt.Sprintf("/tmp/%s", strings.Replace(format.ToString(), "/", "-", -1)), follow)

			if err != nil {
				stream.Bye(w, err)
			}
		}
	}
}
