package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/node"
	"github.com/simplecontainer/smr/pkg/startup"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"syscall"
	"time"
)

func (api *Api) Nodes(c *gin.Context) {
	bytes, err := json.Marshal(api.Cluster.Cluster.Nodes)

	if err != nil {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", err, nil))
	} else {
		c.JSON(http.StatusOK, common.Response(http.StatusOK, "", nil, bytes))
	}
}

func (api *Api) AddNode(c *gin.Context) {
	if !api.Cluster.Started {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", errors.New("cluster is not started"), nil))
		return
	}

	data, err := io.ReadAll(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", err, nil))
		return
	}

	n := node.NewNode()

	err = json.Unmarshal(data, n)

	if err != nil {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", err, nil))
		return
	}

	existing := api.Cluster.Cluster.Find(n)

	if existing == nil {
		n.NodeID = api.Cluster.Cluster.GenerateID()
	} else {
		n.NodeID = existing.NodeID
	}

	var bytes []byte
	bytes, err = n.ToJSON()

	if err != nil {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", err, nil))
		return
	}

	n.ConfChange = raftpb.ConfChange{
		Type:    raftpb.ConfChangeAddNode,
		NodeID:  n.NodeID,
		Context: bytes,
	}

	api.Cluster.NodeConf <- *n

	c.JSON(http.StatusOK, common.Response(http.StatusOK, "node added", nil, bytes))
}

func (api *Api) RemoveNode(c *gin.Context) {
	if !api.Cluster.Started {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", errors.New("cluster is not started"), nil))
		return
	}

	id, err := strconv.ParseUint(c.Param("node"), 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", err, nil))
	}

	n := node.NewNode()
	n.NodeID = id
	n.ConfChange = raftpb.ConfChange{
		Type:   raftpb.ConfChangeRemoveNode,
		NodeID: id,
	}

	api.Cluster.NodeConf <- *n

	c.JSON(http.StatusOK, common.Response(http.StatusOK, "node deleted", nil, nil))
}

func (api *Api) GetNode(c *gin.Context) {
	nodeID, err := api.parseNodeID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "please provide valid node id", err, nil))
		return
	}

	n := api.Cluster.Cluster.FindById(nodeID)
	if n == nil {
		c.JSON(http.StatusNotFound, common.Response(http.StatusNotFound, "node not found", nil, nil))
		return
	}

	bytes, err := n.ToJSON()
	if err != nil {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", err, nil))
		return
	}

	c.JSON(http.StatusOK, common.Response(http.StatusOK, "node found", nil, network.ToJSON(bytes)))
}

func (api *Api) GetNodeVersion(c *gin.Context) {
	nodeID, err := api.parseNodeID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "please provide valid node id", err, nil))
		return
	}

	n := api.Cluster.Cluster.FindById(nodeID)
	if n == nil {
		c.JSON(http.StatusNotFound, common.Response(http.StatusNotFound, "node not found", nil, nil))
		return
	}

	response := network.Send(api.Manager.Http.Clients[api.Manager.User.Username].Http, fmt.Sprintf("%s/version", n.API), http.MethodGet, nil)
	c.JSON(response.HttpStatus, response)
}

func (api *Api) parseNodeID(c *gin.Context) (uint64, error) {
	if c.Param("id") == "" {
		return 0, fmt.Errorf("missing node id")
	}

	nodeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid node id format")
	}

	return nodeID, nil
}

func (api *Api) ListenNode() {
	for {
		select {
		case n, ok := <-api.Cluster.NodeConf:
			if ok {
				switch n.ConfChange.Type {
				case raftpb.ConfChangeAddNode:
					api.Cluster.Cluster.Add(&n)

					api.SaveClusterConfiguration()
					startup.Save(api.Config)

					api.Cluster.Regenerate(api.Config, api.Keys)
					api.Keys.Reloader.ReloadC <- syscall.SIGHUP

					api.Manager.Http, _ = client.GenerateHttpClients(api.Config.NodeName, api.Keys, api.Config.HostPort, api.Cluster)

					api.Cluster.KVStore.ConfChangeC <- n.ConfChange

					logger.Log.Info("added new node")
					break
				case raftpb.ConfChangeRemoveNode:
					nodeID := n.NodeID

					if nodeID != api.Cluster.Node.NodeID {
						api.Cluster.Cluster.Remove(&n)

						api.SaveClusterConfiguration()
						startup.Save(api.Config)

						api.Cluster.Regenerate(api.Config, api.Keys)
						api.Keys.Reloader.ReloadC <- syscall.SIGHUP

						api.Manager.Http, _ = client.GenerateHttpClients(api.Config.NodeName, api.Keys, api.Config.HostPort, api.Cluster)

						api.Cluster.KVStore.ConfChangeC <- n.ConfChange

						logger.Log.Info("removed node from the cluster", zap.Uint64("nodeID", nodeID))
					} else {
						if api.Cluster.RaftNode.IsLeader.Load() {
							logger.Log.Info(fmt.Sprintf("attempt to transfer leader role to %d", api.Cluster.Peers().Nodes[0].NodeID))

							ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
							api.Cluster.RaftNode.TransferLeadership(ctx, api.Cluster.Peers().Nodes[0].NodeID)

							ticker := time.NewTicker(5 * time.Millisecond)
							defer ticker.Stop()

							for {
								if !api.Cluster.RaftNode.IsLeader.Load() {
									break
								}
								select {
								case <-ctx.Done():
									panic("timed out transfering leadership")
								case <-ticker.C:
								}
							}

							logger.Log.Info(fmt.Sprintf("transefered leader role to %d", api.Cluster.Peers().Nodes[0].NodeID))
						}

						go func() {
							<-api.Cluster.RaftNode.Done()

							logger.Log.Info("raft is stopped")
							api.Cluster.NodeFinalizer <- n

							return
						}()

						api.Cluster.KVStore.ConfChangeC <- n.ConfChange
					}
				}
			} else {
				logger.Log.Error("channel for node updates closed")
			}
			break
		}
	}
}
