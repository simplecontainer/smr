package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/control"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/network"
	"io"
	"net/http"
)

func (api *Api) Control(c *gin.Context) {
	data, err := io.ReadAll(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "failed to start upgrade - try again", err, nil))
		return
	}

	batch := &control.CommandBatch{}

	if err = json.Unmarshal(data, &batch); err != nil {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "invalid control data", err, nil))
		return
	}

	if api.Cluster.Node.NodeID == batch.NodeID {
		for _, cmd := range batch.GetCommands() {
			err = cmd.Node(api.Manager, cmd.Data())

			if err != nil {
				c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
				return
			}
		}
	} else {
		target := api.Cluster.Cluster.FindById(batch.NodeID)

		if target == nil {
			c.JSON(http.StatusNotFound, common.Response(http.StatusNotFound, "node not found", nil, nil))
			return
		}

		response := network.Send(
			api.Manager.Http.Clients[api.Manager.User.Username].Http,
			fmt.Sprintf("%s%s", target.API, "/api/v1/cluster/control"),
			http.MethodPost,
			data,
		)

		c.JSON(response.HttpStatus, response)
	}
}
