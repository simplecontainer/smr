package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/kinds/container/shared"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/static"
	"io"
	"net/http"
	"strconv"
)

func (api *Api) Logs(c *gin.Context) {
	group := c.Param("group")
	identifier := c.Param("identifier")
	follow := c.Param("follow")

	w := c.Writer
	header := w.Header()
	header.Set("Transfer-Encoding", "chunked")
	header.Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	container := api.KindsRegistry[static.KIND_CONTAINER].GetShared().(*shared.Shared).Registry.Find(group, identifier)

	if container == nil {
		_, err := w.Write([]byte("container is not found"))

		if err != nil {
			logger.Log.Error(err.Error())
		}

		w.(http.Flusher).Flush()
		w.CloseNotify()
		return
	} else {
		if container.IsGhost() {
			resp, err := network.Raw(api.Manager.Http.Clients[container.GetRuntime().NodeName].Http, fmt.Sprintf("%s/api/v1/logs/%s/%s/%s", api.Manager.Http.Clients[container.GetRuntime().NodeName].API, group, identifier, follow), http.MethodGet, nil)

			var bytes int
			buff := make([]byte, 512)

			for {
				bytes, err = resp.Body.Read(buff)

				if bytes == 0 || err == io.EOF {
					err = resp.Body.Close()

					if err != nil {
						logger.Log.Error(err.Error())
					}

					w.CloseNotify()
					break
				}

				_, err = w.Write(buff[:bytes])

				if err != nil {
					logger.Log.Error(err.Error())
					w.CloseNotify()
					break
				}

				w.(http.Flusher).Flush()
			}
		} else {
			followBool, err := strconv.ParseBool(follow)

			if err != nil {
				_, err = w.Write([]byte(err.Error()))

				if err != nil {
					logger.Log.Error(err.Error())
				}

				w.(http.Flusher).Flush()
				w.CloseNotify()
				return
			}

			reader, err := container.Logs(followBool)

			if err != nil {
				_, err = w.Write([]byte(err.Error()))

				if err != nil {
					logger.Log.Error(err.Error())
				}

				w.(http.Flusher).Flush()
				w.CloseNotify()
				return
			}

			var bytes int
			buff := make([]byte, 512)

			for {
				bytes, err = reader.Read(buff)

				if bytes == 0 || err == io.EOF {
					logger.Log.Info(err.Error())
					w.CloseNotify()
					break
				}

				_, err = w.Write(buff[:bytes])

				if err != nil {
					logger.Log.Error(err.Error())
					w.CloseNotify()
					break
				}

				w.(http.Flusher).Flush()
			}
		}
	}
}
