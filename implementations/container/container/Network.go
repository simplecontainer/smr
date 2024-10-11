package container

import (
	"context"
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	dockerClient "github.com/docker/docker/client"
	"github.com/simplecontainer/smr/pkg/logger"
)

const STARTING_SUBNET string = "10.10.0.0/16"

func (container *Container) AddNetworkInfoTS(networkId string, ipAddress string, networkName string) {
	container.Runtime.NetworkLock.Lock()

	container.Runtime.Networks[networkId] = Network{
		NetworkId:   networkId,
		NetworkName: networkName,
		IP:          ipAddress,
	}

	container.Runtime.NetworkLock.Unlock()
}

func (container *Container) GetNetworkInfoTS() map[string]Network {
	container.Runtime.NetworkLock.RLock()

	networkCopy := make(map[string]Network)
	for k, v := range container.Runtime.Networks {
		networkCopy[k] = v
	}

	container.Runtime.NetworkLock.RUnlock()
	return networkCopy
}

func (container *Container) ConnectToNetwork(containerId string, networkId string) error {
	if c, _ := container.Get(); c != nil && c.State == "running" {
		ctx := context.Background()
		cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}
		defer cli.Close()

		// TODO: Don't connect if the network is same

		EndpointSettings := &network.EndpointSettings{
			NetworkID: networkId,
		}

		err = cli.NetworkConnect(ctx, networkId, containerId, EndpointSettings)

		if err != nil {
			logger.Log.Error(err.Error())
			return errors.New("failed to connect to the network")
		}

		return nil
	} else {
		return errors.New("container is not running")
	}
}

func (container *Container) GetNetwork() *network.NetworkingConfig {
	dnetw := network.NetworkingConfig{EndpointsConfig: map[string]*network.EndpointSettings{}}

	if container.Static.NetworkMode != "host" {
		for _, netw := range container.Static.Networks.Networks {
			dnetw.EndpointsConfig[netw.Reference.Name] = &network.EndpointSettings{
				NetworkID: netw.Reference.Name,
			}
		}
	}

	return &dnetw
}

func (container *Container) FindNetworkAlias(endpointName string, networkId string) bool {
	ctx := context.Background()
	cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	networks, err := cli.NetworkInspect(ctx, networkId, types.NetworkInspectOptions{})

	if err != nil {
		panic(err)
	}

	for _, c := range networks.Containers {
		if c.Name == endpointName {
			return true
		}
	}

	return false
}
