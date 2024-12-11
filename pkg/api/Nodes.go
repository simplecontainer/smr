package api

import (
	"errors"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/static"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"net"
	"net/http"
	"net/url"
	"strings"
)

func (api *Api) AddNode(c *gin.Context) {
	if !api.Cluster.Started {
		c.JSON(http.StatusBadRequest, contracts.ResponseOperator{
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
		c.JSON(http.StatusBadRequest, contracts.ResponseOperator{
			Explanation:      "",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})
	}

	api.Cluster.Cluster.Add(newNode)

	var url *url.URL

	for _, n := range api.Cluster.Cluster.Nodes {
		url, err = client.ParseHostURL(n.URL)

		if err != nil {
			logger.Log.Error(err.Error())
			continue
		}

		tmp := strings.Split(url.Host, ":")

		if net.ParseIP(tmp[0]) != nil {
			api.Config.IPs = append(api.Config.IPs, tmp[0])
		} else {
			api.Config.Domains = append(api.Config.Domains, tmp[0])
		}
	}

	logger.Log.Info("regenerating server certificate to support cluster nodes")

	err = api.Keys.GenerateClient(api.Config.Domains, api.Config.IPs, "root")

	if err != nil {
		logger.Log.Error(err.Error())
		return
	}

	err = api.Keys.GenerateServer(api.Config.Domains, api.Config.IPs)

	if err != nil {
		logger.Log.Error(err.Error())
		return
	}

	err = api.Keys.Clients["root"].Write(static.SMR_SSH_HOME, "root")
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}

	err = api.Keys.Server.Write(static.SMR_SSH_HOME)
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}

	api.Cluster.KVStore.ConfChangeC <- raftpb.ConfChange{
		Type:    raftpb.ConfChangeAddNode,
		NodeID:  newNode.NodeID,
		Context: []byte(newNode.URL),
	}

	api.SaveClusterConfiguration()

	c.JSON(http.StatusOK, contracts.ResponseOperator{
		Explanation:      "",
		ErrorExplanation: "everything went ok",
		Error:            false,
		Success:          true,
		Data:             nil,
	})
}

func (api *Api) RemoveNode(c *gin.Context) {
	if !api.Cluster.Started {
		c.JSON(http.StatusBadRequest, contracts.ResponseOperator{
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
		c.JSON(http.StatusBadRequest, contracts.ResponseOperator{
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

	c.JSON(http.StatusOK, contracts.ResponseOperator{
		Explanation:      "",
		ErrorExplanation: "everything went ok",
		Error:            false,
		Success:          true,
		Data:             nil,
	})
}
