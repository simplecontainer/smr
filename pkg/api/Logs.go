package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/proxy/plain"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/simplecontainer/smr/pkg/stream"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
)

func (a *Api) Logs(c *gin.Context) {
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

	containerShared, ok := a.KindsRegistry[static.KIND_CONTAINERS].GetShared().(*shared.Shared)
	if !ok {
		stream.ByeWithStatus(w, http.StatusBadRequest, errors.New("container registry not available"))
		return
	}

	container := containerShared.Registry.Find(static.SMR_PREFIX, group, name)
	if container == nil {
		stream.ByeWithStatus(w, http.StatusBadRequest, errors.New(fmt.Sprintf("%s '%s/%s' not found", static.KIND_CONTAINERS, group, name)))
		return
	}

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer func() {
		logger.Log.Info("proxy context cancel called")
		cancel()
	}()

	if container.IsGhost() {
		client, ok := a.Manager.Http.Clients[container.GetRuntime().Node.NodeName]
		if !ok {
			stream.ByeWithStatus(w, http.StatusBadRequest, errors.New(fmt.Sprintf("node for %s '%s/%s' not found", static.KIND_CONTAINERS, group, name)))
			return
		}

		URL := fmt.Sprintf("%s/api/v1/logs/%s/%s/%s", client.API, format.ToString(), which, c.Param("follow"))

		var remote io.ReadCloser
		remote, err = plain.Dial(ctx, cancel, client.Http, URL)

		if err != nil {
			stream.ByeWithStatus(w, http.StatusBadRequest, err)
			return
		}

		proxy(ctx, cancel, c, remote)
	} else {
		switch which {
		case "main", "init":
			var reader io.ReadCloser

			if which == "init" {
				reader, err = container.GetInit().Logs(ctx, follow)
			} else {
				reader, err = container.Logs(ctx, follow)
			}

			if err != nil {
				stream.ByeWithStatus(w, http.StatusBadRequest, err)
				return
			}

			proxy(ctx, cancel, c, reader)
		default:
			stream.ByeWithStatus(w, http.StatusBadRequest, errors.New("container can be only main or init"))
		}
	}
}

func proxy(ctx context.Context, cancel context.CancelFunc, c *gin.Context, remote io.ReadCloser) {
	proxy := plain.Create(ctx, cancel, c.Writer, remote)

	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write([]byte{0, 0, 0, 0, 0, 0, 0, 0})
	c.Writer.Flush()

	err := proxy.Proxy()

	if err != nil {
		logger.Log.Error("proxy returned error", zap.Error(err))
	}

	stream.Bye(c.Writer, nil)
}
