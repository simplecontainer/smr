package api

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/cluster"
	"github.com/simplecontainer/smr/pkg/controler"
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
	"syscall"
)

var starting = false

func (api *Api) StartCluster(c *gin.Context) {
	if starting {
		c.JSON(http.StatusConflict, common.Response(http.StatusConflict, "", errors.New("cluster is in the process of starting on this node"), nil))
		return
	} else {
		if api.Cluster != nil && api.Cluster.Started {
			c.JSON(http.StatusConflict, common.Response(http.StatusConflict, "", errors.New("cluster already started"), nil))
			return
		}
	}

	starting = true
	data, err := io.ReadAll(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
		return
	}

	var control *controler.Control
	err = json.Unmarshal(data, &control)

	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
		return
	}

	var parsed *url.URL
	parsed, err = url.Parse(control.GetStart().NodeRaftAPI)

	if err != nil {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", err, nil))
	}

	api.Cluster, err = cluster.Restore(api.Config)
	peers := node.NewNodes()

	if err != nil {
		api.Cluster = cluster.New()
		api.Cluster.Node = api.Cluster.Cluster.NewNode(api.Config.NodeName, control.GetStart().NodeRaftAPI, fmt.Sprintf("https://%s:%s", parsed.Hostname(), api.Config.HostPort.Port))
		api.Cluster.Node.Version = api.Version

		api.Cluster.Cluster.Add(api.Cluster.Node)

		if api.Config.KVStore.Peer != "" {
			peerNode := node.NewNode()
			peerNode.API = api.Config.KVStore.Peer
			peers.Add(peerNode)
		}
	} else {
		api.Cluster.Node.Version = api.Version
		peers = api.Cluster.Peers()
	}

	user := &authentication.User{}

	if api.Config.KVStore.Join {
		for _, peer := range peers.Nodes {
			// Find any valid certificate for the domain or ip
			clientObj := api.Manager.Http.FindValidFor(peer.API)

			if clientObj != nil {
				user = authentication.New(clientObj.Username, clientObj.API)
			}

			if user.Username == "" {
				continue
			}

			data, err = json.Marshal(api.Cluster.Node)

			if err != nil {
				c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", errors.New("user not found for remote agent"), nil))
				return
			}

			response := network.Send(api.Manager.Http.Clients[user.Username].Http, fmt.Sprintf("%s/api/v1/cluster/node", peer.API), http.MethodPost, data)

			if response.Error {
				c.JSON(http.StatusBadRequest, response)
				return
			}

			err = json.Unmarshal(response.Data, &api.Cluster.Node)

			if err != nil {
				c.JSON(http.StatusBadRequest, response)
				return
			}

			response = network.Send(api.Manager.Http.Clients[user.Username].Http, fmt.Sprintf("%s/api/v1/cluster/", peer.API), http.MethodGet, nil)

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
					api.Cluster.Cluster.AddOrUpdate(n)
				}
			} else {
				c.JSON(http.StatusBadRequest, response)
				return
			}
		}

		if user.Username == "" {
			c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", errors.New("user not found for remote agent"), nil))
			return
		}
	}

	api.Manager.Cluster = api.Cluster

	CAPool := x509.NewCertPool()
	CAPool.AddCert(api.Keys.CA.Certificate)

	tlsConfig := &tls.Config{
		ClientAuth:     tls.RequireAndVerifyClientCert,
		ClientCAs:      CAPool,
		GetCertificate: api.Keys.Reloader.GetCertificateFunc(),
	}

	api.Cluster.Regenerate(api.Config, api.Keys)
	api.Keys.Reloader.ReloadC <- syscall.SIGHUP
	api.Manager.Http, err = client.GenerateHttpClients(api.Config.NodeName, api.Keys, api.Config.HostPort, api.Cluster)

	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
		return
	}

	err = api.SetupCluster(tlsConfig, api.Cluster.Node, api.Cluster, api.Config.KVStore.Join)

	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
		return
	}

	api.SaveClusterConfiguration()

	go func() {
		select {
		case <-api.Cluster.InSync:
			// Replay after RAFT synced with cluster

			_, err = api.Manager.KindsRegistry[static.KIND_CONTAINERS].Replay(api.Manager.User)

			if err != nil {
				logger.Log.Error("failed to replay containers", zap.Error(err))
			}

			_, err = api.Manager.KindsRegistry[static.KIND_GITOPS].Replay(api.Manager.User)

			if err != nil {
				logger.Log.Error("failed to replay gitops", zap.Error(err))
			}
			break
		}
	}()

	go events.Listen(api.Manager.KindsRegistry, api.Replication.EventsC, api.Replication.Informer, api.Wss)
	go api.ListenNode()
	go api.Replication.ListenData(api.Config.NodeName)

	err = flannel.Setup(c, api.Etcd, control.GetStart().Overlay, control.GetStart().Backend)

	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", errors.New("flannel overlay network failed to start"), nil))
		return
	}

	api.Cluster.Started = true
	api.Cluster.Node.Version = api.Version

	event, err := events.NewNodeEvent(events.EVENT_CONTROL_SUCCESS, api.Cluster.Node)

	if err != nil {
		logger.Log.Error("failed to dispatch node event", zap.Error(err))
	} else {
		logger.Log.Info("dispatched node event", zap.String("event", event.GetType()))
		events.Dispatch(event, api.KindsRegistry[static.KIND_NODE].GetShared().(*shared.Shared), api.Cluster.Node.NodeID)
	}

	c.JSON(http.StatusOK, common.Response(http.StatusOK, "cluster started on this node", nil, network.ToJSON(map[string]string{
		"name": api.Config.NodeName,
	})))

	return
}

func (api *Api) GetCluster(c *gin.Context) {
	c.JSON(http.StatusOK, common.Response(http.StatusOK, "cluster started", nil, network.ToJSON(api.Cluster.Cluster.Nodes)))
}

func (api *Api) SaveClusterConfiguration() {
	api.Config.KVStore.Cluster = api.Cluster.Cluster.Nodes
	api.Config.KVStore.Node = api.Cluster.Node
	api.Config.KVStore.URL = api.Cluster.Node.URL
	api.Config.KVStore.API = api.Cluster.Node.API
	api.Config.KVStore.Join = true

	err := startup.Save(api.Config)
	if err != nil {
		logger.Log.Error(err.Error())
	}

	if api.Cluster.Node != nil {
		format := f.New(static.SMR_PREFIX, static.CATEGORY_PLAIN, "cluster", "internal", "cluster")
		obj := objects.New(api.Manager.Http.Clients[api.User.Username], api.User)

		var bytes []byte
		var err error
		bytes, err = json.Marshal(api.Cluster.Cluster.Nodes)

		if err == nil {
			obj.Propose(format, bytes)
		} else {
			logger.Log.Error(err.Error())
		}
	}
}
