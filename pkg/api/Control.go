package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/controler"
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	cshared "github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/shared"
	nshared "github.com/simplecontainer/smr/pkg/kinds/node/shared"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/static"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

func (api *Api) Control(c *gin.Context) {
	const (
		drainTimeout  = 160 * time.Second
		pollInterval  = 500 * time.Millisecond
		controlAPIURL = "/api/v1/cluster/control"
	)

	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "failed to start upgrade - try again", err, nil))
		return
	}

	var control controler.Control
	if err := json.Unmarshal(data, &control); err != nil {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "invalid control data", err, nil))
		return
	}

	valid, err := control.Validate()
	if err != nil || !valid {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "invalid control content", err, nil))
		return
	}

	if !api.Cluster.Started {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "cluster is not started yet", err, nil))
		return
	}

	if api.Cluster.Node.NodeID == control.Drain.NodeID {
		api.Manager.KindsRegistry[static.KIND_GITOPS].GetShared().(*shared.Shared).Watchers.Drain()
		api.Manager.KindsRegistry[static.KIND_CONTAINERS].GetShared().(*cshared.Shared).Watchers.Drain()

		go api.handleDrainProcess(control, data, c)

		c.JSON(http.StatusOK, common.Response(http.StatusOK, "process of upgrading the node started", nil, nil))
		return
	}

	target := api.Cluster.Cluster.FindById(control.Drain.NodeID)
	if target == nil {
		c.JSON(http.StatusNotFound, common.Response(http.StatusNotFound, "node not found", nil, nil))
		return
	}

	response := network.Send(
		api.Manager.Http.Clients[api.Manager.User.Username].Http,
		fmt.Sprintf("%s%s", target.API, controlAPIURL),
		http.MethodPost,
		data,
	)

	c.JSON(response.HttpStatus, response)
}

func (api *Api) handleDrainProcess(control controler.Control, data []byte, c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 160*time.Second)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Error("draining timeout exceeded 160 seconds, drain aborted - manual intervention needed")
			return

		case <-ticker.C:
			if api.isAllWatchersDrained() {
				// TODO: Configure state of the node for all cases
				ticker.Stop()
				api.Cluster.Node.SetDrain(true)

				if control.GetUpgrade() != nil {
					api.Cluster.Node.SetUpgrade(true)
				}

				event, err := events.NewNodeEvent(events.EVENT_CONTROL_START, api.Cluster.Node)

				if err != nil {
					logger.Log.Error("failed to dispatch node event", zap.Error(err))
				} else {
					logger.Log.Info("dispatched node event", zap.String("event", event.GetType()))

					eventStr := event.ToFormat().ToString()
					api.Replication.Informer.AddCh(eventStr)
					events.Dispatch(event, api.KindsRegistry[static.KIND_NODE].GetShared().(*nshared.Shared), api.Cluster.Node.NodeID)

					select {
					case <-api.Replication.Informer.GetCh(eventStr):
						api.Replication.Informer.RmCh(eventStr)
					case <-ctx.Done():
						logger.Log.Warn("timed out waiting for event acknowledgment")
					}
				}

				api.Cluster.Node.ConfChange = raftpb.ConfChange{
					Type:    raftpb.ConfChangeRemoveNode,
					NodeID:  control.Drain.NodeID,
					Context: data,
				}

				api.Cluster.NodeConf <- *api.Cluster.Node
			}
			break

		case finalized := <-api.Cluster.NodeFinalizer:
			var finalControl controler.Control

			if err := json.Unmarshal(finalized.ConfChange.Context, &finalControl); err != nil {
				logger.Log.Info("invalid finalizer context", zap.Error(err))
				continue
			}

			if finalControl.Timestamp == control.Timestamp {
				logger.Log.Info("finalizing node", zap.Uint64("node", finalized.NodeID))

				if err := control.Apply(c, api.Etcd); err != nil {
					logger.Log.Error("control process start error", zap.Error(err))
				}

				return
			} else {
				logger.Log.Error("timestamp mismatch in finalizer")
			}
			break
		}
	}
}

func (api *Api) isAllWatchersDrained() bool {
	gitops := api.Manager.KindsRegistry[static.KIND_GITOPS].GetShared().(*shared.Shared)
	containers := api.Manager.KindsRegistry[static.KIND_CONTAINERS].GetShared().(*cshared.Shared)

	return len(gitops.Watchers.Repositories) == 0 && len(containers.Watchers.Watchers) == 0
}
