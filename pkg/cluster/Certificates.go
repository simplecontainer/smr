package cluster

import (
	"fmt"
	"github.com/docker/docker/client"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/static"
	"net"
	"net/url"
	"strings"
)

func (cluster *Cluster) Regenerate(config *configuration.Configuration, keys *keys.Keys) {
	var url *url.URL
	var err error

	for _, n := range cluster.Cluster.Nodes {
		url, err = client.ParseHostURL(n.URL)

		if err != nil {
			logger.Log.Error(err.Error())
			continue
		}

		tmp := strings.Split(url.Host, ":")

		if net.ParseIP(tmp[0]) != nil {
			config.Certificates.IPs.Add(tmp[0])
		} else {
			config.Certificates.Domains.Add(tmp[0])
		}
	}

	logger.Log.Info("regenerating server certificate to support cluster nodes")

	err = keys.GenerateClient(config.Certificates.Domains, config.Certificates.IPs, config.NodeName)

	if err != nil {
		logger.Log.Error(err.Error())
		return
	}

	err = keys.GenerateServer(config.Certificates.Domains, config.Certificates.IPs)

	if err != nil {
		logger.Log.Error(err.Error())
		return
	}

	err = keys.Clients[config.NodeName].Write(static.SMR_SSH_HOME, config.NodeName)
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}

	err = keys.Server.Write(static.SMR_SSH_HOME, config.NodeName)
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}

	fmt.Println("regenerated certificates")
}
