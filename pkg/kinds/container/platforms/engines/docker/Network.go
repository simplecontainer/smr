package docker

import (
	"context"
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	IDClient "github.com/docker/docker/client"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/engines/docker/internal"
)

func (container *Docker) UpdateNetworkInfoTS(networkId string, ipAddress string, networkName string) {
	container.Networks.Lock.Lock()

	network := container.Networks.Find(networkId)

	if network != nil {
		network.Docker.IP = ipAddress
		network.Docker.NetworkId = networkId
	}

	container.Networks.Lock.Unlock()
}

func (container *Docker) RemoveNetworkInfoTS(containerId string, networkId string, ipAddress string, networkName string) error {
	container.Networks.Lock.Lock()

	err := container.Networks.Remove(containerId, networkId)

	container.Networks.Lock.Unlock()

	if err != nil {
		return err
	}

	return nil
}

func (container *Docker) GetNetworkInfoTS() *internal.Networks {
	container.Networks.Lock.RLock()

	networks := container.Networks

	container.Networks.Lock.RUnlock()
	return networks
}

func (container *Docker) SyncNetworkInformation() error {
	ctx := context.Background()
	cli, err := IDClient.NewClientWithOpts(IDClient.FromEnv, IDClient.WithAPIVersionNegotiation())

	if err != nil {
		panic(err)
	}

	defer func(cli *IDClient.Client) {
		err = cli.Close()
		if err != nil {
			return
		}
	}(cli)

	if container.DockerID != "" {
		data, _ := cli.ContainerInspect(ctx, container.DockerID)
		var networkInspect types.NetworkResource

		if data.NetworkSettings == nil {
			return errors.New("network settings empty")
		}

		for _, dockerNetwork := range data.NetworkSettings.Networks {
			networkInspect, err = cli.NetworkInspect(ctx, dockerNetwork.NetworkID, types.NetworkInspectOptions{
				Scope:   "",
				Verbose: false,
			})

			if err != nil {
				return err
			}

			if container.Networks.Find(networkInspect.ID) != nil {
				container.UpdateNetworkInfoTS(networkInspect.ID, dockerNetwork.IPAddress, networkInspect.Name)
			} else {
				// optimistic
				container.RemoveNetworkInfoTS(container.DockerID, networkInspect.ID, dockerNetwork.IPAddress, networkInspect.Name)
			}
		}

		return nil
	}

	return errors.New("docker id is not set")
}

func (container *Docker) BuildNetwork() *network.NetworkingConfig {
	networks := network.NetworkingConfig{EndpointsConfig: map[string]*network.EndpointSettings{}}

	if container.NetworkMode != "host" {
		for _, netw := range container.Networks.Networks {
			networks.EndpointsConfig[netw.Reference.Name] = &network.EndpointSettings{
				NetworkID: netw.Docker.NetworkId,
			}
		}
	}

	return &networks
}
