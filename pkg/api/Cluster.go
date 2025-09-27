package api

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/cluster"
	"github.com/simplecontainer/smr/pkg/contracts/icontrol"
	"github.com/simplecontainer/smr/pkg/control"
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/flannel"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/node/shared"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/node"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/startup"
	"github.com/simplecontainer/smr/pkg/static"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"sync"
	"syscall"
)

var lock = &sync.RWMutex{}

func (a *Api) StartCluster(c *gin.Context) {
	lock.Lock()
	defer lock.Unlock()

	if a.Cluster != nil && a.Cluster.Started {
		c.JSON(http.StatusConflict, common.Response(http.StatusConflict, "", errors.New(static.CLUSTER_STARTED), nil))
		return
	}

	data, err := io.ReadAll(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
		return
	}

	var batch control.CommandBatch

	err = json.Unmarshal(data, &batch)

	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
		return
	}

	var cmd icontrol.Command
	cmd, err = batch.GetCommand("start")

	if err != nil {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", err, nil))
		return
	}

	var parsed *url.URL
	parsed, err = url.Parse(cmd.Data()["raft"])

	if err != nil {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", err, nil))
		return
	}

	a.Cluster, err = cluster.Restore(a.Config)
	peers := node.NewNodes()

	if err != nil {
		a.Cluster = cluster.New()
		a.Cluster.Node = a.Cluster.Cluster.NewNode(a.Config.NodeName, parsed.String(), fmt.Sprintf("https://%s:%s", parsed.Hostname(), a.Config.HostPort.Port))
		a.Cluster.Node.Version = a.Version

		a.Cluster.Cluster.Add(a.Cluster.Node)

		if a.Config.KVStore.Peer != "" {
			peerNode := node.NewNode()
			peerNode.API = a.Config.KVStore.Peer
			peers.Add(peerNode)
		}
	} else {
		a.Cluster.Node.State.ResetControl()
		a.Cluster.Node.Version = a.Version
		peers = a.Cluster.Peers()
	}

	user := &authentication.User{}

	if a.Cluster.Join || a.Config.KVStore.Join {
		for _, peer := range peers.Nodes {
			// Find any valid certificate for the domain or ip
			clientObj := a.Manager.Http.FindValidFor(peer.API)

			if clientObj != nil {
				user = authentication.New(clientObj.Username, clientObj.API)
			}

			if user.Username == "" {
				continue
			}

			data, err = json.Marshal(a.Cluster.Node)
			if err != nil {
				c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", errors.New(static.USER_NOT_FOUND), nil))
				return
			}

			response := network.Send(a.Manager.Http.Clients[user.Username].Http, fmt.Sprintf("%s/api/v1/cluster/node", peer.API), http.MethodPost, data)
			if response.Error {
				c.JSON(http.StatusBadRequest, response)
				return
			}

			err = json.Unmarshal(response.Data, &a.Cluster.Node)
			if err != nil {
				c.JSON(http.StatusBadRequest, response)
				return
			}

			response = network.Send(a.Manager.Http.Clients[user.Username].Http, fmt.Sprintf("%s/api/v1/cluster/", peer.API), http.MethodGet, nil)
			if response.Success {
				var bytes []byte
				var tmp []*node.Node

				bytes, err = json.Marshal(response.Data)

				if err != nil {
					c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
					return
				}

				err = json.Unmarshal(bytes, &tmp)

				if err != nil {
					c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
					return
				}

				for _, n := range tmp {
					a.Cluster.Cluster.AddOrUpdate(n)
				}
			} else {
				c.JSON(http.StatusBadRequest, response)
				return
			}
		}

		if user.Username == "" {
			c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", errors.New(static.USER_NOT_FOUND), nil))
			return
		}
	}

	a.Manager.Cluster = a.Cluster

	CAPool := x509.NewCertPool()
	CAPool.AddCert(a.Keys.CA.Certificate)

	tlsConfig := &tls.Config{
		ClientAuth:     tls.RequireAndVerifyClientCert,
		ClientCAs:      CAPool,
		GetCertificate: a.Keys.Reloader.GetCertificateFunc(),
	}

	a.Cluster.Regenerate(a.Config, a.Keys)
	a.Keys.Reloader.ReloadC <- syscall.SIGHUP
	a.Manager.Http, err = clients.GenerateHttpClients(a.Keys, a.Config.HostPort, a.Cluster)

	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
		return
	}

	a.SetupReplication()
	go a.Replication.ListenData(a.Config.NodeName)

	err = a.SetupCluster(tlsConfig, a.Cluster.Node, a.Cluster, a.Config.KVStore.Join)

	if err != nil {
		logger.Log.Error("failed to setup clusters", zap.Error(err))
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
		return
	}

	a.SaveClusterConfiguration()

	go a.ListenNode()
	go events.Listen(a.Manager.KindsRegistry, a.Replication.EventsC, a.Replication.Informer, a.Wss)

	err = flannel.Setup(c, a.Etcd, cmd.Data()["cidr"], cmd.Data()["backend"])

	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", errors.New(static.FLANNEL_START_FAILED), nil))
		return
	}

	a.Cluster.Started = true
	a.Cluster.Node.Version = a.Version

	c.JSON(http.StatusOK, common.Response(http.StatusOK, static.CLUSTER_STARTED_OK, nil, network.ToJSON(map[string]string{
		"name": a.Config.NodeName,
	})))

	go func() {
		event, err := events.NewNodeEvent(events.EVENT_CLUSTER_STARTED, a.Cluster.Node)

		if err != nil {
			logger.Log.Error("failed to dispatch node event", zap.Error(err))
		} else {
			logger.Log.Info("dispatched node event", zap.String("event", event.GetType()))
			events.Dispatch(event, a.KindsRegistry[static.KIND_NODE].GetShared().(*shared.Shared), a.Cluster.Node.NodeID)
		}
	}()

	return
}

func (a *Api) StatusCluster(c *gin.Context) {
	c.JSON(http.StatusOK, common.Response(http.StatusOK, static.CLUSTER_STARTED_OK, nil, network.ToJSON(a.Cluster.Cluster.Nodes)))
}

func (a *Api) SaveClusterConfiguration() {
	a.Config.KVStore.Cluster = a.Cluster.Cluster.Nodes
	a.Config.KVStore.Node = a.Cluster.Node
	a.Config.KVStore.URL = a.Cluster.Node.URL
	a.Config.KVStore.API = a.Cluster.Node.API
	a.Config.KVStore.Replay = true

	// After later restarts/upgrades node needs to join the cluster
	// This behavior is only desired in the multi node cluster - standalone node ignore
	if len(a.Cluster.Cluster.Nodes) > 1 {
		a.Config.KVStore.Join = true
	}

	err := startup.Save(a.Config, a.Config.Environment.Container, 0)
	if err != nil {
		logger.Log.Error(err.Error())
	}

	if a.Cluster.Node != nil {
		format := f.New(static.SMR_PREFIX, static.CATEGORY_PLAIN, "cluster", "internal", "cluster")
		obj := objects.New(a.Manager.Http.Clients[a.User.Username], a.User)

		var bytes []byte
		bytes, err = json.Marshal(a.Cluster.Cluster.Nodes)

		if err == nil {
			obj.Propose(format, bytes)
		} else {
			logger.Log.Error(err.Error())
		}
	}
}
