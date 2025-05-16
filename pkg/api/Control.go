package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/control"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/network"
	"go.uber.org/zap"
	"io"
	"net/http"
)

func (a *Api) Control(c *gin.Context) {
	data, err := io.ReadAll(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusInternalServerError, "failed to read control data", err, nil))
		return
	}

	batch := control.NewCommandBatch()

	if err = json.Unmarshal(data, &batch); err != nil {
		logger.Log.Error("err", zap.Error(err))
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "invalid control data", err, nil))
		return
	}

	if a.Cluster.Node.NodeID == batch.GetNodeID() {
		go func() {
			for _, cmd := range batch.GetCommands() {
				err = cmd.Node(a, cmd.Data())

				if err != nil {
					logger.Log.Error("error running control on the node", zap.Error(err))
				}
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err = batch.Put(ctx, a.Etcd)

			if err != nil {
				logger.Log.Error("failed to inform client about control event", zap.Error(err))
			}
		}()

		c.JSON(http.StatusOK, common.Response(http.StatusOK, "controls batch applied", nil, nil))
	} else {
		target := a.Cluster.Cluster.FindById(batch.GetNodeID())

		if target == nil {
			c.JSON(http.StatusNotFound, common.Response(http.StatusNotFound, "node not found", nil, nil))
			return
		}

		response := network.Send(
			a.Manager.Http.Clients[a.Manager.User.Username].Http,
			fmt.Sprintf("%s%s", target.API, "/api/v1/cluster/control"),
			http.MethodPost,
			data,
		)

		c.JSON(response.HttpStatus, response)
	}
}
