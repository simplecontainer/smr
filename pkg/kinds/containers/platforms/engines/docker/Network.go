package docker

import (
	TDNetwork "github.com/docker/docker/api/types/network"
	IDClient "github.com/docker/docker/client"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/engines/docker/internal"
)

func (container *Docker) SyncNetwork() error {
	containerInspected, err := internal.Inspect(container.DockerID)

	if err != nil {
		if IDClient.IsErrNotFound(err) {
			for _, network := range container.Networks.Networks {
				container.RemoveNetworkInfo(container.DockerID, network.Docker.NetworkId, network.Docker.IP, network.Reference.Name)
			}

			return nil
		}

		return err
	}

	var networkInspected TDNetwork.Inspect

	if containerInspected.NetworkSettings == nil {
		return nil
	}

	for _, network := range containerInspected.NetworkSettings.Networks {
		networkInspected, err = internal.InspectNetwork(network.NetworkID)

		if err != nil {
			return err
		}

		if container.Networks.Find(networkInspected.ID) != nil {
			container.UpdateNetworkInfo(networkInspected.ID, network.IPAddress, networkInspected.Name)
		} else {
			container.RemoveNetworkInfo(container.DockerID, networkInspected.ID, network.IPAddress, networkInspected.Name)
		}
	}

	return nil
}

func (container *Docker) UpdateNetworkInfo(networkId string, ipAddress string, networkName string) {
	network := container.Networks.Find(networkId)

	if network != nil {
		network.Docker.IP = ipAddress
		network.Docker.NetworkId = networkId
	}
}

func (container *Docker) RemoveNetworkInfo(containerId string, networkId string, ipAddress string, networkName string) error {
	err := container.Networks.Remove(containerId, networkId)

	if err != nil {
		return err
	}

	return nil
}

func (container *Docker) BuildNetwork() *TDNetwork.NetworkingConfig {
	networks := TDNetwork.NetworkingConfig{EndpointsConfig: map[string]*TDNetwork.EndpointSettings{}}

	if container.NetworkMode != "host" {
		for _, netw := range container.Networks.Networks {
			networks.EndpointsConfig[netw.Reference.Name] = &TDNetwork.EndpointSettings{
				NetworkID: netw.Docker.NetworkId,
			}
		}
	}

	return &networks
}
