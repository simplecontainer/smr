package container

import (
	"github.com/docker/docker/api/types/network"
	"github.com/simplecontainer/smr/implementations/container/container/internal"
)

func (container *Container) AddNetworkInfoTS(networkId string, ipAddress string, networkName string) {
	container.Runtime.NetworkLock.Lock()

	network := container.Runtime.Networks.Find(networkId)

	if network != nil {
		network.Docker.IP = ipAddress
	}

	container.Runtime.NetworkLock.Unlock()
}

func (container *Container) GetNetworkInfoTS() *internal.Networks {
	container.Runtime.NetworkLock.RLock()

	networks := container.Runtime.Networks

	container.Runtime.NetworkLock.RUnlock()
	return networks
}

func (container *Container) GetNetwork() *network.NetworkingConfig {
	dnetw := network.NetworkingConfig{EndpointsConfig: map[string]*network.EndpointSettings{}}

	if container.Static.NetworkMode != "host" {
		for _, netw := range container.Static.Networks.Networks {
			dnetw.EndpointsConfig[netw.Reference.Name] = &network.EndpointSettings{
				NetworkID: netw.Docker.NetworkId,
			}
		}
	}

	return &dnetw
}
