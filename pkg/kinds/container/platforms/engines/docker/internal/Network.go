package internal

import (
	"context"
	"errors"
	"github.com/docker/docker/api/types"
	dockerNetwork "github.com/docker/docker/api/types/network"
	dockerClient "github.com/docker/docker/client"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"sync"
)

type Networks struct {
	Networks []*Network
	Lock     sync.RWMutex
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

func NewNetworks(networks []v1.ContainerNetwork) *Networks {
	networksObj := &Networks{
		Networks: make([]*Network, 0),
	}

	if len(networks) == 0 {
		networksObj.Add(v1.ContainerNetwork{
			Group: "docker",
			Name:  "bridge",
		})
	}

	for _, network := range networks {
		networksObj.Add(network)
	}

	networksObj.Add(v1.ContainerNetwork{
		Group: "docker",
		Name:  "flannel",
	})

	return networksObj
}

func NewNetwork(network v1.ContainerNetwork) *Network {
	return &Network{
		Reference: NetworkReference{
			Group: network.Group,
			Name:  network.Name,
		},
		Docker: NetworkDocker{
			NetworkId: GetNetworkId(network.Name),
			IP:        "",
		},
	}
}

func (networks *Networks) Add(network v1.ContainerNetwork) {
	networks.Networks = append(networks.Networks, NewNetwork(network))
}

func (networks *Networks) Remove(containerId string, networkId string) error {
	for i, n := range networks.Networks {
		if n.Docker.NetworkId == networkId {
			err := n.Disconnect(containerId)

			if err != nil {
				return err
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

func GetNetworkId(name string) string {
	ctx := context.Background()
	cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	var networks []types.NetworkResource
	networks, err = cli.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		panic(err)
	}

	for _, c := range networks {
		if c.Name == name {
			return c.ID
		}
	}

	return ""
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

func (network *Network) FindNetworkAlias(endpointName string) error {
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

	networks, err := cli.NetworkInspect(ctx, network.Docker.NetworkId, types.NetworkInspectOptions{})

	if err != nil {
		return err
	}

	for _, c := range networks.Containers {
		if c.Name == endpointName {
			return errors.New("endpoint already exists")
		}
	}

	return nil
}
