package api

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/cluster"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/networking"
	"github.com/simplecontainer/smr/pkg/node"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/startup"
	"github.com/simplecontainer/smr/pkg/static"
	"go.etcd.io/etcd/server/v3/embed"
	"io"
	"log"
	"net/http"
	"net/url"
	"syscall"
	"time"
)

var starting = false

func (api *Api) StartCluster(c *gin.Context) {
	if starting {
		c.JSON(http.StatusConflict, contracts.Response{
			Explanation:      "",
			ErrorExplanation: "cluster is in the process of starting on this node",
			Error:            true,
			Success:          false,
			Data: network.ToJson(map[string]any{
				"name": api.Config.NodeName,
			}),
		})

		return
	} else {
		if api.Cluster != nil && api.Cluster.Started {
			c.JSON(http.StatusConflict, contracts.Response{
				Explanation:      "",
				ErrorExplanation: "cluster is already started on this node",
				Error:            true,
				Success:          false,
				Data: network.ToJson(map[string]any{
					"name": api.Config.NodeName,
				}),
			})

			return
		}
	}

	starting = true
	data, err := io.ReadAll(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusInternalServerError, contracts.Response{
			Explanation:      "",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})

		return
	}

	var request map[string]string
	err = json.Unmarshal(data, &request)

	if err != nil {
		c.JSON(http.StatusInternalServerError, contracts.Response{
			Explanation:      "",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})

		return
	}

	api.Cluster, err = cluster.Restore(api.Config)

	if err != nil {
		api.Cluster = cluster.New()
	}

	currentNode := api.Cluster.Cluster.NewNode(request["node"])

	if request["join"] != "" {
		user := &authentication.User{}

		var URL *url.URL
		URL, err = url.Parse(request["join"])

		for _, client := range api.Manager.Http.Clients {
			for _, domain := range client.Domains {
				if domain == URL.Hostname() {
					user = &authentication.User{
						Username: client.Username,
						Domain:   domain,
					}
				}
			}

			for _, ip := range client.IPs {
				if ip.String() == URL.Hostname() {
					user = &authentication.User{
						Username: client.Username,
						Domain:   ip.String(),
					}
				}
			}
		}

		if user == nil {
			c.JSON(http.StatusBadRequest, contracts.Response{
				HttpStatus:       http.StatusBadRequest,
				Explanation:      "user not found for remote agent",
				ErrorExplanation: "",
				Error:            false,
				Success:          false,
				Data:             nil,
			})

			return
		}

		d, _ := json.Marshal(map[string]string{"node": request["node"]})
		response := network.Send(api.Manager.Http.Clients[user.Username].Http, fmt.Sprintf("%s/api/v1/cluster/node", request["join"]), http.MethodPost, d)

		if response.Error {
			c.JSON(http.StatusBadRequest, response)
			return
		}

		// Ask join: what is the cluster?
		response = network.Send(api.Manager.Http.Clients[user.Username].Http, fmt.Sprintf("%s/api/v1/cluster", request["join"]), http.MethodGet, nil)

		if response.Success {
			var bytes []byte
			var peers map[string][]*node.Node

			bytes, err = json.Marshal(response.Data)

			if err != nil {
				c.JSON(http.StatusInternalServerError, contracts.Response{
					Explanation:      "",
					ErrorExplanation: err.Error(),
					Error:            true,
					Success:          false,
					Data:             nil,
				})

				return
			}

			err = json.Unmarshal(bytes, &peers)

			if err != nil {
				c.JSON(http.StatusInternalServerError, contracts.Response{
					Explanation:      "",
					ErrorExplanation: err.Error(),
					Error:            true,
					Success:          false,
					Data:             nil,
				})

				return
			}

			for _, n := range peers["cluster"] {
				api.Cluster.Cluster.Add(n)

				if n.URL == currentNode.URL {
					api.Cluster.Node = n
				}
			}
		} else {
			c.JSON(http.StatusBadRequest, response)
			return
		}
	} else {
		// If not joining generate node yourself
		api.Cluster.Node = currentNode
		api.Cluster.Cluster.Add(currentNode)
	}

	api.Manager.Cluster = api.Cluster

	var server *embed.Etcd
	server, err = networking.StartEtcd(api.Config)

	if err != nil {
		c.JSON(http.StatusInternalServerError, contracts.Response{
			Explanation:      "",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})

		return
	}

	select {
	case <-server.Server.ReadyNotify():
		fmt.Println("etcd server started - continue with starting raft")

		CAPool := x509.NewCertPool()
		CAPool.AddCert(api.Keys.CA.Certificate)

		tlsConfig := &tls.Config{
			ClientAuth:     tls.RequireAndVerifyClientCert,
			ClientCAs:      CAPool,
			GetCertificate: api.Keys.Reloader.GetCertificateFunc(),
		}

		api.Cluster.Regenerate(api.Config, api.Keys)
		api.Keys.Reloader.ReloadC <- syscall.SIGHUP

		err = api.SetupKVStore(tlsConfig, api.Cluster.Node.NodeID, api.Cluster, request["join"])

		if err != nil {
			server.Server.Stop()

			c.JSON(http.StatusInternalServerError, contracts.Response{
				Explanation:      "",
				ErrorExplanation: err.Error(),
				Error:            true,
				Success:          false,
				Data:             nil,
			})

			return
		}

		api.Config.KVStore.Node = api.Cluster.Node.NodeID
		api.Config.KVStore.URL = api.Cluster.Node.URL
		api.Config.KVStore.Cluster = api.Cluster.Cluster.Nodes
		api.Config.KVStore.JoinCluster = request["join"] != ""

		api.SaveClusterConfiguration()

		go api.Replication.ListenEtcd(api.Config.NodeName)
		go api.Replication.ListenData(api.Config.NodeName)

		err = networking.Flannel(api.Config.OverlayNetwork)

		if err != nil {
			c.JSON(http.StatusInternalServerError, contracts.Response{
				Explanation:      "",
				ErrorExplanation: "flannel overlay network failed to start",
				Error:            true,
				Success:          false,
				Data:             nil,
			})

			return
		}

		api.Cluster.Started = true

		c.JSON(http.StatusOK, contracts.Response{
			Explanation:      "",
			ErrorExplanation: "everything went ok",
			Error:            false,
			Success:          true,
			Data: network.ToJson(map[string]string{
				"name": api.Config.NodeName,
			}),
		})

		return
	case <-time.After(60 * time.Second):
		server.Server.Stop() // trigger a shutdown
		log.Printf("etcd server took too long to start!")

		c.JSON(http.StatusInternalServerError, contracts.Response{
			Explanation:      "",
			ErrorExplanation: "etcd server took too long to start",
			Error:            true,
			Success:          false,
			Data:             nil,
		})
	}
}

func (api *Api) GetCluster(c *gin.Context) {
	c.JSON(http.StatusOK, contracts.Response{
		Explanation:      "",
		ErrorExplanation: "list of peers",
		Error:            false,
		Success:          true,
		Data: network.ToJson(map[string]any{
			"cluster": api.Cluster.Cluster.Nodes,
		}),
	})
}

func (api *Api) SaveClusterConfiguration() {
	err := startup.Save(api.Config)
	if err != nil {
		logger.Log.Error(err.Error())
	}

	format := f.NewUnformated("smr.cluster", static.CATEGORY_PLAIN_STRING)
	obj := objects.New(api.Manager.Http.Clients[api.User.Username], api.User)

	bytes, err := json.Marshal(api.Cluster.Cluster.Nodes)

	if err == nil {
		obj.Propose(format, bytes)
	} else {
		logger.Log.Error(err.Error())
	}
}
