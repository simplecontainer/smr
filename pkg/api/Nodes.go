package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/clients"
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

func (a *Api) Nodes(c *gin.Context) {
	bytes, err := json.Marshal(a.Cluster.Cluster.Nodes)

	if err != nil {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", err, nil))
	} else {
		c.JSON(http.StatusOK, common.Response(http.StatusOK, "", nil, bytes))
	}
}

func (a *Api) AddNode(c *gin.Context) {
	if !a.Cluster.Started {
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

	existing := a.Cluster.Cluster.Find(n)

	if existing == nil {
		n.NodeID = a.Cluster.Cluster.GenerateID()
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

	a.Cluster.NodeConf <- *n

	c.JSON(http.StatusOK, common.Response(http.StatusOK, "node added", nil, bytes))
}
func (a *Api) RemoveNode(c *gin.Context) {
	if !a.Cluster.Started {
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

	a.Cluster.NodeConf <- *n

	c.JSON(http.StatusOK, common.Response(http.StatusOK, "node deleted", nil, nil))
}

func (a *Api) GetNode(c *gin.Context) {
	nodeID, err := a.parseNodeID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "please provide valid node id", err, nil))
		return
	}

	n := a.Cluster.Cluster.FindById(nodeID)
	if n == nil {
		c.JSON(http.StatusNotFound, common.Response(http.StatusNotFound, "node not found", nil, nil))
		return
	}

	bytes, err := n.ToJSON()
	if err != nil {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", err, nil))
		return
	}

	c.JSON(http.StatusOK, common.Response(http.StatusOK, "node found", nil, bytes))
}

func (a *Api) GetNodeVersion(c *gin.Context) {
	nodeID, err := a.parseNodeID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "please provide valid node id", err, nil))
		return
	}

	n := a.Cluster.Cluster.FindById(nodeID)
	if n == nil {
		c.JSON(http.StatusNotFound, common.Response(http.StatusNotFound, "node not found", nil, nil))
		return
	}

	response := network.Send(a.Manager.Http.Clients[a.Manager.User.Username].Http, fmt.Sprintf("%s/version", n.API), http.MethodGet, nil)
	c.JSON(response.HttpStatus, response)
}

func (a *Api) parseNodeID(c *gin.Context) (uint64, error) {
	if c.Param("id") == "" {
		return 0, fmt.Errorf("missing node id")
	}

	nodeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid node id format")
	}

	return nodeID, nil
}

func (a *Api) ListenNode() {
	for {
		select {
		case n, ok := <-a.Cluster.NodeConf:
			if ok {
				switch n.ConfChange.Type {
				case raftpb.ConfChangeAddNode:
					a.Cluster.Cluster.Add(&n)

					a.SaveClusterConfiguration()
					startup.Save(a.Config, a.Config.Environment.Container, 0)

					a.Cluster.Regenerate(a.Config, a.Keys)
					a.Keys.Reloader.ReloadC <- syscall.SIGHUP

					a.Manager.Http, _ = clients.GenerateHttpClients(a.Keys, a.Config.HostPort, a.Cluster)

					a.Cluster.KVStore.ConfChangeC <- n.ConfChange

					logger.Log.Info("added new node")
					break
				case raftpb.ConfChangeRemoveNode:
					nodeID := n.NodeID

					if nodeID != a.Cluster.Node.NodeID {
						timeout := 5 * time.Second
						ticker := time.NewTicker(500 * time.Millisecond)

						timeoutChan := time.After(timeout)
						pool := true

						for pool {
							select {
							case <-timeoutChan:
								logger.Log.Error("timed out waiting for the peer to be removed from RAFT")
								pool = false
								break
							case <-ticker.C:
								if peer := a.Cluster.Cluster.FindById(nodeID); peer == nil {
									ticker.Stop()
									a.Cluster.Cluster.Remove(&n)

									a.SaveClusterConfiguration()
									startup.Save(a.Config, a.Config.Environment.Container, 0)

									a.Cluster.Regenerate(a.Config, a.Keys)
									a.Keys.Reloader.ReloadC <- syscall.SIGHUP

									a.Manager.Http, _ = clients.GenerateHttpClients(a.Keys, a.Config.HostPort, a.Cluster)

									logger.Log.Info("removed node from the cluster", zap.Uint64("nodeID", nodeID))

									pool = false
									break
								}

								logger.Log.Info("wait for node to be removed from the cluster", zap.Uint64("nodeID", nodeID))
							}
						}

						ticker.Stop()
						break
					} else {
						if len(a.Cluster.Peers().Nodes) > 0 && a.Cluster.Node.NodeID != a.Cluster.Peers().Nodes[0].NodeID {
							if a.Cluster.RaftNode.IsLeader.Load() {
								logger.Log.Info(fmt.Sprintf("attempt to transfer leader role to %d", a.Cluster.Peers().Nodes[0].NodeID))

								ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
								a.Cluster.RaftNode.TransferLeadership(ctx, a.Cluster.Peers().Nodes[0].NodeID)

								ticker := time.NewTicker(5 * time.Millisecond)
								defer ticker.Stop()

								for {
									if !a.Cluster.RaftNode.IsLeader.Load() {
										break
									}
									select {
									case <-ctx.Done():
										logger.Log.Error("timed out waiting for the peer to transfer lead")
										return
									case <-ticker.C:
									}
								}

								logger.Log.Info(fmt.Sprintf("transefered leader role to %d", a.Cluster.Peers().Nodes[0].NodeID))
							}
						}

						go func() {
							<-a.Cluster.RaftNode.Done()

							logger.Log.Info("raft is stopped")
							a.Cluster.NodeFinalizer <- n

							return
						}()

						if len(a.Cluster.Peers().Nodes) > 0 && a.Cluster.Node.NodeID != a.Cluster.Peers().Nodes[0].NodeID {
							a.Cluster.KVStore.ConfChangeC <- n.ConfChange
						} else {
							a.Cluster.NodeFinalizer <- n
						}
					}
				}
			} else {
				logger.Log.Error("channel for node updates closed")
			}
			break
		}
	}
}
