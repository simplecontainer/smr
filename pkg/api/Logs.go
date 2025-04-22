package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/simplecontainer/smr/pkg/stream"
)

func (api *Api) Logs(c *gin.Context) {
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
	w.WriteHeader(http.StatusOK)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Minute)
	defer cancel()

	c.Request = c.Request.WithContext(ctx)

	containerShared, ok := api.KindsRegistry[static.KIND_CONTAINERS].GetShared().(*shared.Shared)
	if !ok {
		stream.Bye(w, errors.New("container registry not available"))
		return
	}

	container := containerShared.Registry.Find(static.SMR_PREFIX, group, name)
	if container == nil {
		stream.Bye(w, errors.New(fmt.Sprintf("%s '%s/%s' not found", static.KIND_CONTAINER, group, name)))
		return
	}

	if container.IsGhost() {
		client, ok := api.Manager.Http.Clients[container.GetRuntime().Node.NodeName]
		if !ok {
			stream.Bye(w, errors.New(fmt.Sprintf("node for %s '%s/%s' not found", static.KIND_CONTAINER, group, name)))
			return
		}

		err := stream.StreamRemote(c, w, fmt.Sprintf("%s/api/v1/logs/%s/%s/%s", client.API, format.ToString(), which, c.Param("follow")), client)

		if err != nil {
			stream.Bye(w, err)
		}

		return
	}

	switch which {
	case "main":
		handleContainerLogs(w, container, follow, false)
	case "init":
		handleContainerLogs(w, container, follow, true)
	default:
		stream.Bye(w, errors.New("container can be only main or init"))
	}
}

func handleContainerLogs(w gin.ResponseWriter, container platforms.IContainer, follow bool, isInit bool) {
	var reader io.ReadCloser
	var err error

	if isInit {
		reader, err = container.GetInit().Logs(follow)
	} else {
		reader, err = container.Logs(follow)
	}

	if err != nil {
		stream.Bye(w, err)
		return
	}

	if reader == nil {
		stream.Bye(w, errors.New("no logs available"))
		return
	}

	defer reader.Close()

	err = stream.Stream(w, reader)
	if err != nil {
		stream.Bye(w, err)
	}
}
