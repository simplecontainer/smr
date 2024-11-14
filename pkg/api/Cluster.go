package api

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/cluster"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/startup"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/spf13/viper"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"go.etcd.io/etcd/server/v3/embed"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func (api *Api) StartCluster(c *gin.Context) {
	CAPool := x509.NewCertPool()
	CAPool.AddCert(api.Keys.CA.Certificate)

	var PEMCertificate []byte
	var PEMPrivateKey []byte

	var err error

	PEMCertificate, err = keys.PEMEncode(keys.CERTIFICATE, api.Keys.Server.CertificateBytes)
	PEMPrivateKey, err = keys.PEMEncode(keys.PRIVATE_KEY, api.Keys.Server.PrivateKeyBytes)

	serverTLSCert, err := tls.X509KeyPair(PEMCertificate, PEMPrivateKey)
	if err != nil {
		logger.Log.Fatal("error opening certificate and key file for control connection", zap.String("error", err.Error()))
		return
	}

	tlsConfig := &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    CAPool,
		Certificates: []tls.Certificate{serverTLSCert},
	}

	if viper.GetBool("restore") {
		api.Cluster, err = cluster.Restore(api.Config)
	} else {
		api.Cluster, err = cluster.New(c.Request.Body)
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, contracts.ResponseOperator{
			Explanation:      "",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})

		return
	}

	var server *embed.Etcd
	server, err = api.Cluster.StartSingleNodeEtcd(api.Config)

	if err != nil {
		c.JSON(http.StatusInternalServerError, contracts.ResponseOperator{
			Explanation:      "",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data: map[string]any{
				"cluster": api.Cluster.Cluster,
			},
		})

		return
	}

	select {
	case <-server.Server.ReadyNotify():
		api.Cluster.EtcdClient = cluster.NewEtcdClient()
		api.SetupKVStore(tlsConfig, api.Cluster.Node.NodeID, api.Cluster.Cluster)

		go api.Cluster.ListenEvents()

		deadline := time.Now().Add(10 * time.Second)
		for {
			if api.Cluster.RaftNode.NodeReady() != nil || time.Now().After(deadline) {
				c.JSON(http.StatusInternalServerError, contracts.ResponseOperator{
					Explanation:      "",
					ErrorExplanation: "flannel overlay network failed to start",
					Error:            true,
					Success:          false,
					Data:             nil,
				})
				return
			}
		}

		select {
		case <-api.Cluster.RaftNode.NodeReady():
			err = api.Cluster.ConfigureFlannel(api.Config.OverlayNetwork)

			if err != nil {
				c.JSON(http.StatusInternalServerError, contracts.ResponseOperator{
					Explanation:      "",
					ErrorExplanation: "flannel overlay network failed to start",
					Error:            true,
					Success:          false,
					Data:             nil,
				})
				return
			}
			break
		}

		c.JSON(http.StatusOK, contracts.ResponseOperator{
			Explanation:      "",
			ErrorExplanation: "everything went ok",
			Error:            false,
			Success:          true,
			Data:             nil,
		})
	case <-time.After(60 * time.Second):
		server.Server.Stop() // trigger a shutdown
		log.Printf("Server took too long to start!")

		c.JSON(http.StatusInternalServerError, contracts.ResponseOperator{
			Explanation:      "",
			ErrorExplanation: "etcd server took too long to start",
			Error:            true,
			Success:          false,
			Data:             nil,
		})
	}
}

func (api *Api) GetCluster(c *gin.Context) {
	c.JSON(http.StatusOK, contracts.ResponseOperator{
		Explanation:      "",
		ErrorExplanation: "list of peers",
		Error:            false,
		Success:          true,
		Data: map[string]any{
			"cluster": api.Cluster.Cluster,
		},
	})
}

func (api *Api) EtcdPut(c *gin.Context) {
	timeout, err := time.ParseDuration("20s")

	if err != nil {
		c.JSON(http.StatusBadRequest, contracts.ResponseOperator{
			Explanation:      "",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})

		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	var body []byte
	body, err = io.ReadAll(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusBadRequest, contracts.ResponseOperator{
			Explanation:      "",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})

		cancel()
		return
	}

	_, err = api.Cluster.EtcdClient.Put(ctx, c.Param("key"), string(body))
	cancel()

	c.JSON(http.StatusOK, contracts.ResponseOperator{
		Explanation:      "",
		ErrorExplanation: "all goodies",
		Error:            false,
		Success:          true,
		Data: map[string]any{
			"cluster": api.Cluster.Cluster,
		},
	})
}
func (api *Api) EtcdDelete(c *gin.Context) {}

func (api *Api) AddNode(c *gin.Context) {
	node, err := cluster.NewNodeRequest(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusBadRequest, contracts.ResponseOperator{
			Explanation:      "",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})
	}

	api.Cluster.Add(node)
	api.Cluster.KVStore.ConfChangeC <- raftpb.ConfChange{
		Type:    raftpb.ConfChangeAddNode,
		NodeID:  node.NodeID,
		Context: []byte(node.URL),
	}
}
func (api *Api) RemoveNode(c *gin.Context) {
	node, err := cluster.NewNodeRequest(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusBadRequest, contracts.ResponseOperator{
			Explanation:      "",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})
	}

	api.Cluster.Remove(node)

	api.Cluster.KVStore.ConfChangeC <- raftpb.ConfChange{
		Type:   raftpb.ConfChangeRemoveNode,
		NodeID: node.NodeID,
	}
}

func (api *Api) XXL() {
	api.Config.KVStore.Node = api.Cluster.Node.NodeID
	api.Config.KVStore.URL = api.Cluster.Node.URL
	api.Config.KVStore.Cluster = api.Cluster.Cluster

	fmt.Println(api.Config)

	var out io.Writer
	var err error

	out, err = os.OpenFile(fmt.Sprintf("%s/%s/config.yaml", api.Config.Environment.PROJECTDIR, static.CONFIGDIR), os.O_WRONLY, 0644)

	err = startup.Save(api.Config, out)
	if err != nil {
		logger.Log.Error(err.Error())
	}
}
