package internal

import (
	"context"
	"errors"
	dockerNetwork "github.com/docker/docker/api/types/network"
	dockerClient "github.com/docker/docker/client"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/logger"
)

type Networks struct {
	Networks []*Network
}

type Network struct {
	Reference NetworkReference
	Docker    NetworkDocker
}

type NetworkReference struct {
	Group string
	Name  string
}

type NetworkDocker struct {
	NetworkId string
	IP        string
}

func NewNetworks(networks []v1.ContainersNetwork) *Networks {
	networksObj := &Networks{
		Networks: make([]*Network, 0),
	}

	var bridgeFound = false

	for _, network := range networks {
		networksObj.Add(network)

		if network.Name == "bridge" {
			bridgeFound = true
		}
	}

	if !bridgeFound {
		networksObj.Add(v1.ContainersNetwork{
			Group: "docker",
			Name:  "bridge",
		})
	}

	return networksObj
}

func NewNetwork(network v1.ContainersNetwork) *Network {
	nw, err := GetNetwork(network.Name)

	if err != nil {
		return nil
	}

	return &Network{
		Reference: NetworkReference{
			Group: network.Group,
			Name:  network.Name,
		},
		Docker: NetworkDocker{
			NetworkId: nw.ID,
			IP:        "",
		},
	}
}

func (networks *Networks) Add(network v1.ContainersNetwork) {
	networks.Networks = append(networks.Networks, NewNetwork(network))
}

func (networks *Networks) Remove(containerId string, networkId string) error {
	for i, n := range networks.Networks {
		if n.Docker.NetworkId == networkId {
			err := n.Disconnect(containerId)

			if err != nil {
				logger.Log.Error(err.Error())
			}

			networks.Networks = append(networks.Networks[:i], networks.Networks[i+1:]...)
			return nil
		}
	}

	return errors.New("network not found in container network registry")
}

func (networks *Networks) Find(networkId string) *Network {
	for i, n := range networks.Networks {
		if n.Docker.NetworkId == networkId {
			return networks.Networks[i]
		}
	}

	return nil
}

func (network *Network) Connect(containerId string) error {
	ctx := context.Background()
	cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())

	if err != nil {
		panic(err)
	}

	defer func(cli *dockerClient.Client) {
		err = cli.Close()
		if err != nil {
			return
		}
	}(cli)

	EndpointSettings := &dockerNetwork.EndpointSettings{
		NetworkID: network.Docker.NetworkId,
	}

	err = cli.NetworkConnect(ctx, network.Docker.NetworkId, containerId, EndpointSettings)

	if err != nil {
		return err
	}

	return nil
}

func (network *Network) Disconnect(containerId string) error {
	ctx := context.Background()
	cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())

	if err != nil {
		panic(err)
	}

	defer func(cli *dockerClient.Client) {
		err = cli.Close()
		if err != nil {
			return
		}
	}(cli)

	err = cli.NetworkDisconnect(ctx, network.Docker.NetworkId, containerId, true)

	if err != nil {
		return err
	}

	return nil
}

func (network *Network) EndpointExists(endpointName string) bool {
	inspected, err := InspectNetwork(network.Docker.NetworkId)

	if err != nil {
		return false
	}

	for _, c := range inspected.Containers {
		if c.Name == endpointName {
			return true
		}
	}

	return false
}
