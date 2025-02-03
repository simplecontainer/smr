package api

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/container/shared"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/simplecontainer/smr/pkg/stream"
	"io"
	"net/http"
	"strconv"
)

func (api *Api) Logs(c *gin.Context) {
	prefix := c.Param("prefix")
	version := c.Param("version")
	category := c.Param("category")
	kind := c.Param("kind")
	group := c.Param("group")
	name := c.Param("name")

	follow, err := strconv.ParseBool(c.Param("follow"))

	if err != nil {
		follow = false
	}

	format := f.New(prefix, version, category, kind, group, name)

	w := c.Writer
	header := w.Header()
	header.Set("Transfer-Encoding", "chunked")
	header.Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	container := api.KindsRegistry[static.KIND_CONTAINER].GetShared().(*shared.Shared).Registry.Find(static.SMR_PREFIX, group, name)

	if container == nil {
		stream.Bye(w, errors.New(fmt.Sprintf("%s is not found", static.KIND_CONTAINER)))
		return
	} else {
		if container.IsGhost() {
			client, ok := api.Manager.Http.Clients[container.GetRuntime().NodeName]

			if !ok {
				stream.Bye(w, errors.New(fmt.Sprintf("%s is not found", static.KIND_CONTAINER)))
				return
			} else {
				stream.StreamRemote(w, fmt.Sprintf("https://%s/api/v1/logs/%s/%s", client.API, format.ToString(), c.Param("follow")), client)
			}
		} else {
			var reader io.ReadCloser
			reader, err = container.Logs(follow)

			if err != nil {
				network.StreamByte([]byte(err.Error()), w)
			}

			err = stream.Stream(w, reader)

			if err != nil {
				stream.Bye(w, err)
			}
		}
	}
}
