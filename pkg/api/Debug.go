package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/proxy/plain"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/simplecontainer/smr/pkg/stream"
	"github.com/simplecontainer/smr/pkg/tail"
	"go.uber.org/zap"
	"io"
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

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

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

				var remote io.ReadCloser
				remote, err = plain.Dial(ctx, cancel, client.Http, URL)

				if err != nil {
					stream.ByeWithStatus(w, http.StatusBadRequest, err)
					return
				}

				proxy := plain.Create(ctx, cancel, c.Writer, remote)

				w.WriteHeader(http.StatusOK)
				err = proxy.Proxy()

				if err != nil {
					logger.Log.Error("proxy returned error", zap.Error(err))
				}

				stream.Bye(w, nil)
			} else {
				var reader io.ReadCloser

				PATH := fmt.Sprintf("/tmp/%s", strings.Replace(format.ToString(), "/", "-", -1))
				reader, err = tail.File(c.Request.Context(), PATH, follow)

				if err != nil {
					stream.ByeWithStatus(w, http.StatusBadRequest, err)
					return
				}

				proxy := plain.Create(ctx, cancel, c.Writer, reader)

				// Say hello back to open connection
				c.Writer.WriteHeader(http.StatusOK)
				c.Writer.Write([]byte{0, 0, 0, 0, 0, 0, 0, 0})
				c.Writer.Flush()

				err = proxy.Proxy()

				if err != nil {
					logger.Log.Error("proxy returned error", zap.Error(err))
				}

				stream.Bye(w, nil)
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

				var remote io.ReadCloser
				remote, err = plain.Dial(ctx, cancel, client.Http, URL)

				if err != nil {
					stream.ByeWithStatus(w, http.StatusBadRequest, err)
					return
				}

				proxy := plain.Create(ctx, cancel, c.Writer, remote)

				w.WriteHeader(http.StatusOK)
				err = proxy.Proxy()

				if err != nil {
					logger.Log.Error("proxy returned error", zap.Error(err))
				}

				stream.Bye(w, nil)
			}
		} else {
			var reader io.ReadCloser

			PATH := fmt.Sprintf("/tmp/%s", strings.Replace(format.ToString(), "/", "-", -1))
			reader, err = tail.File(c.Request.Context(), PATH, follow)

			if err != nil {
				stream.ByeWithStatus(w, http.StatusBadRequest, err)
				return
			}

			proxy := plain.Create(ctx, cancel, c.Writer, reader)

			// Say hello back to open connection
			c.Writer.WriteHeader(http.StatusOK)
			c.Writer.Write([]byte{0, 0, 0, 0, 0, 0, 0, 0})
			c.Writer.Flush()

			err = proxy.Proxy()

			if err != nil {
				logger.Log.Error("proxy returned error", zap.Error(err))
			}

			stream.Bye(w, nil)
		}
	}
}
