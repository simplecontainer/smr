package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/kinds/container/shared"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/static"
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

	container := api.KindsRegistry[static.KIND_CONTAINER].GetShared().(*shared.Shared).Registry.Find(static.SMR_PREFIX, group, identifier)

	if container == nil {
		err := network.StreamByte([]byte("container is not found"), w)

		if err != nil {
			logger.Log.Error(err.Error())
		}
	} else {
		if container.IsGhost() {
			client, ok := api.Manager.Http.Clients[container.GetRuntime().NodeName]

			if !ok {
				err := network.StreamByte([]byte("container is not found"), w)

				if err != nil {
					logger.Log.Error(err.Error())
				}
			}

			resp, err := network.Raw(client.Http, fmt.Sprintf("https://%s/api/v1/logs/%s/%s/%s", api.Manager.Http.Clients[container.GetRuntime().NodeName].API, group, identifier, follow), http.MethodGet, nil)

			if err != nil {
				err = network.StreamByte([]byte(err.Error()), w)

				if err != nil {
					logger.Log.Error(err.Error())
				}
			}

			err = network.StreamHttp(resp.Body, w)

			if err != nil {
				err = network.StreamByte([]byte(err.Error()), w)

				if err != nil {
					logger.Log.Error(err.Error())
				}
			}

			network.StreamClose(w)
		} else {
			followBool, err := strconv.ParseBool(follow)

			if err != nil {
				err = network.StreamByte([]byte(err.Error()), w)

				if err != nil {
					logger.Log.Error(err.Error())
				}
			}

			reader, err := container.Logs(followBool)

			if err != nil {
				err = network.StreamByte([]byte(err.Error()), w)

				if err != nil {
					logger.Log.Error(err.Error())
				}
			}

			err = network.StreamHttp(reader, w)

			if err != nil {
				err = network.StreamByte([]byte(err.Error()), w)

				if err != nil {
					logger.Log.Error(err.Error())
				}
			}

			network.StreamClose(w)
		}
	}
}
