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
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/networking"
	"github.com/simplecontainer/smr/pkg/node"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/startup"
	"github.com/simplecontainer/smr/pkg/static"
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

	var request map[string]string
	err = json.Unmarshal(data, &request)

	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
		return
	}

	var parsed *url.URL
	parsed, err = url.Parse(request["node"])

	if err != nil {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", err, nil))
	}

	api.Cluster, err = cluster.Restore(api.Config)

	if err != nil {
		api.Cluster = cluster.New()
		thisNode := api.Cluster.Cluster.NewNode(api.Config.NodeName, request["node"], fmt.Sprintf("%s:%s", parsed.Hostname(), api.Config.HostPort.Port))

		if request["join"] != "" {
			user := &authentication.User{}

			// Find any valid certificate for the domain or ip
			clientObj := api.Manager.Http.FindValidFor(request["join"])

			if clientObj != nil {
				user = authentication.New(clientObj.Username, clientObj.API)
			}

			if user.Username == "" {
				c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", errors.New("user not found for remote agent"), nil))
				return
			}

			data, _ := json.Marshal(map[string]string{
				"node":     request["node"],
				"nodeName": api.Config.NodeName,
				"API":      fmt.Sprintf("%s:%s", parsed.Hostname(), api.Config.HostPort.Port),
			})

			response := network.Send(api.Manager.Http.Clients[user.Username].Http, fmt.Sprintf("%s/api/v1/cluster/node", request["join"]), http.MethodPost, data)

			if response.Error {
				c.JSON(http.StatusBadRequest, response)
				return
			}

			err = json.Unmarshal(response.Data, &thisNode)

			if err != nil {
				c.JSON(http.StatusBadRequest, response)
				return
			}

			// Ask join: what is the cluster?
			response = network.Send(api.Manager.Http.Clients[user.Username].Http, fmt.Sprintf("%s/api/v1/cluster", request["join"]), http.MethodGet, nil)

			if response.Success {
				var bytes []byte
				var peers []*node.Node

				bytes, err = json.Marshal(response.Data)

				if err != nil {
					c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
					return
				}

				err = json.Unmarshal(bytes, &peers)

				if err != nil {
					c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
					return
				}

				for _, n := range peers {
					api.Cluster.Cluster.Add(n)

					if n.URL == thisNode.URL {
						api.Cluster.Node = n
					}
				}
			} else {
				c.JSON(http.StatusBadRequest, response)
				return
			}
		}

		api.Cluster.Node = thisNode
		api.Cluster.Cluster.Add(thisNode)
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
	api.Manager.Http, err = client.GenerateHttpClients(api.Config.NodeName, api.Keys, api.Cluster)

	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
		return
	}

	err = api.SetupCluster(tlsConfig, api.Cluster.Node.NodeID, api.Cluster, request["join"])

	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
		return
	}

	api.Config.KVStore.Node = api.Cluster.Node.NodeID
	api.Config.KVStore.URL = api.Cluster.Node.URL
	api.Config.KVStore.Cluster = api.Cluster.Cluster.Nodes
	api.Config.KVStore.JoinCluster = request["join"] != ""

	api.SaveClusterConfiguration()

	go events.Listen(api.Manager.KindsRegistry, api.Replication.EventsC)

	go api.ListenNode()
	go api.Replication.ListenOutside(api.Config.NodeName)
	go api.Replication.ListenData(api.Config.NodeName)

	err = networking.Flannel(request["overlay"], request["backend"])

	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", errors.New("flannel overlay network failed to start"), nil))
		return
	}

	api.Cluster.Started = true

	c.JSON(http.StatusOK, common.Response(http.StatusOK, "cluster started on this node", nil, network.ToJson(map[string]string{
		"name": api.Config.NodeName,
	})))
	return
}

func (api *Api) GetCluster(c *gin.Context) {
	c.JSON(http.StatusOK, common.Response(http.StatusOK, "cluster starte", nil, network.ToJson(api.Cluster.Cluster.Nodes)))
}

func (api *Api) SaveClusterConfiguration() {
	err := startup.Save(api.Config)
	if err != nil {
		logger.Log.Error(err.Error())
	}

	format := f.New(static.SMR_PREFIX, static.CATEGORY_PLAIN, "cluster", "internal", "cluster")
	obj := objects.New(api.Manager.Http.Clients[api.User.Username], api.User)

	bytes, err := json.Marshal(api.Cluster.Cluster.Nodes)

	if err == nil {
		obj.Propose(format, bytes)
	} else {
		logger.Log.Error(err.Error())
	}
}
