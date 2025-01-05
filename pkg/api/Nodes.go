package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/contracts"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"net/http"
	"syscall"
)

func (api *Api) AddNode(c *gin.Context) {
	if !api.Cluster.Started {
		c.JSON(http.StatusBadRequest, contracts.Response{
			Explanation:      "",
			ErrorExplanation: errors.New("cluster is not started").Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})

		return
	}

	newNode, err := api.Cluster.Cluster.NewNodeRequest(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusBadRequest, contracts.Response{
			Explanation:      "",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})
	}

	api.Cluster.Cluster.Add(newNode)

	api.Cluster.Regenerate(api.Config, api.Keys)
	api.Keys.Reloader.ReloadC <- syscall.SIGHUP

	api.Cluster.KVStore.ConfChangeC <- raftpb.ConfChange{
		Type:    raftpb.ConfChangeAddNode,
		NodeID:  newNode.NodeID,
		Context: []byte(newNode.URL),
	}

	api.Config.KVStore.Node = api.Cluster.Node.NodeID
	api.Config.KVStore.URL = api.Cluster.Node.URL
	api.Config.KVStore.Cluster = api.Cluster.Cluster.Nodes
	api.SaveClusterConfiguration()

	c.JSON(http.StatusOK, contracts.Response{
		Explanation:      "",
		ErrorExplanation: "everything went ok",
		Error:            false,
		Success:          true,
		Data:             nil,
	})
}

func (api *Api) RemoveNode(c *gin.Context) {
	if !api.Cluster.Started {
		c.JSON(http.StatusBadRequest, contracts.Response{
			Explanation:      "",
			ErrorExplanation: errors.New("cluster is not started").Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})

		return
	}

	n, err := api.Cluster.Cluster.NewNodeRequest(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusBadRequest, contracts.Response{
			Explanation:      "",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})
	}

	api.Cluster.Cluster.Remove(n)

	api.Cluster.KVStore.ConfChangeC <- raftpb.ConfChange{
		Type:   raftpb.ConfChangeRemoveNode,
		NodeID: n.NodeID,
	}

	api.SaveClusterConfiguration()

	c.JSON(http.StatusOK, contracts.Response{
		Explanation:      "",
		ErrorExplanation: "everything went ok",
		Error:            false,
		Success:          true,
		Data:             nil,
	})
}
