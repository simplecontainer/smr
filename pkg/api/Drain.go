package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/controler"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	cshared "github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/shared"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/static"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"io"
	"net/http"
	"strconv"
	"time"
)

func (api *Api) Drain(c *gin.Context) {
	data, err := io.ReadAll(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "failed to start upgrade - try again", err, nil))
		return
	}

	var control *controler.Control

	err = json.Unmarshal(data, &control)

	if err != nil {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "invalid node sent", err, nil))
		return
	}

	if api.Cluster.Node.NodeID == control.Drain.NodeID {
		c.AddParam("node", strconv.FormatUint(api.Cluster.Node.NodeID, 10))

		api.Cluster.Node.SetDrain(true)
		api.Manager.KindsRegistry[static.KIND_GITOPS].GetShared().(*shared.Shared).Watchers.Drain()
		api.Manager.KindsRegistry[static.KIND_CONTAINERS].GetShared().(*cshared.Shared).Watchers.Drain()

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 160*time.Second)
			defer cancel()

			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					logger.Log.Error("draining timeout exceeded 160 seconds, drain aborted - could leave incosistent state")
					return

				case <-ticker.C:
					gitopsWatchers := api.Manager.KindsRegistry[static.KIND_GITOPS].GetShared()
					containersWatchers := api.Manager.KindsRegistry[static.KIND_CONTAINERS].GetShared()

					var gitopsEmpty, containersEmpty bool

					if gitopsShared, ok := gitopsWatchers.(*shared.Shared); ok && gitopsShared != nil {
						gitopsEmpty = len(gitopsShared.Watchers.Repositories) == 0
					}

					if containersShared, ok := containersWatchers.(*cshared.Shared); ok && containersShared != nil {
						containersEmpty = len(containersShared.Watchers.Watchers) == 0
					}

					if gitopsEmpty && containersEmpty {
						api.Cluster.KVStore.ConfChangeC <- raftpb.ConfChange{
							Type:   raftpb.ConfChangeRemoveNode,
							NodeID: control.Drain.NodeID,
						}

						return
					}
				}
			}
		}()

		c.JSON(http.StatusOK, common.Response(http.StatusOK, "process of draining the node started", nil, nil))
	} else {
		n := api.Cluster.Cluster.FindById(control.Drain.NodeID)

		if n == nil {
			c.JSON(http.StatusNotFound, common.Response(http.StatusNotFound, "node not found", nil, nil))
		} else {
			response := network.Send(api.Manager.Http.Clients[api.Manager.User.Username].Http, fmt.Sprintf("https://%s/api/v1/drain", n.API), http.MethodPost, data)
			c.JSON(response.HttpStatus, response)
		}
	}
}
