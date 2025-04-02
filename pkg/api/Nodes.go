package api

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/logger"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"go.uber.org/zap"
	"net/http"
	"os"
	"strconv"
	"syscall"
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

	newNode, err := api.Cluster.Cluster.NewNodeRequest(c.Request.Body, api.Cluster.Node.NodeID)

	if err != nil {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", err, nil))
		return
	}

	var bytes []byte
	bytes, err = newNode.ToJson()

	if err != nil {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", err, nil))
		return
	}

	api.Cluster.KVStore.ConfChangeC <- raftpb.ConfChange{
		Type:    raftpb.ConfChangeAddNode,
		NodeID:  newNode.NodeID,
		Context: bytes,
	}

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

	api.Cluster.KVStore.ConfChangeC <- raftpb.ConfChange{
		Type:   raftpb.ConfChangeRemoveNode,
		NodeID: id,
	}

	c.JSON(http.StatusOK, common.Response(http.StatusOK, "node deleted", nil, nil))
}

func (api *Api) ListenNode() {
	for {
		select {
		case n, ok := <-api.Cluster.NodeConf:
			if ok {
				switch n.ConfChange.Type {
				case raftpb.ConfChangeAddNode:
					api.Cluster.Cluster.Add(&n)

					//api.Cluster.Regenerate(api.Config, api.Keys)
					//api.Keys.Reloader.ReloadC <- syscall.SIGHUP

					api.Config.KVStore.Node = api.Cluster.Node.NodeID
					api.Config.KVStore.URL = api.Cluster.Node.URL
					api.Config.KVStore.Cluster = api.Cluster.Cluster.Nodes
					api.SaveClusterConfiguration()

					api.Cluster.Regenerate(api.Config, api.Keys)
					api.Keys.Reloader.ReloadC <- syscall.SIGHUP

					api.Manager.Http, _ = client.GenerateHttpClients(api.Config.NodeName, api.Keys, api.Cluster)

					logger.Log.Info("added new node")
					break
				case raftpb.ConfChangeRemoveNode:
					api.Cluster.Cluster.Remove(&n)

					api.Config.KVStore.Node = api.Cluster.Node.NodeID
					api.Config.KVStore.URL = api.Cluster.Node.URL
					api.Config.KVStore.Cluster = api.Cluster.Cluster.Nodes
					api.SaveClusterConfiguration()

					api.Cluster.Regenerate(api.Config, api.Keys)
					api.Keys.Reloader.ReloadC <- syscall.SIGHUP

					api.Manager.Http, _ = client.GenerateHttpClients(api.Config.NodeName, api.Keys, api.Cluster)

					logger.Log.Info("removed node from the cluster", zap.String("node", n.NodeName))

					if n.NodeID == api.Cluster.Node.NodeID {
						logger.Log.Info("that node is me - proceed with shutdown")
						os.Exit(0)
					}

					break
				}
			} else {
				logger.Log.Error("channel for node updates closed")
			}
			break
		}
	}
}
