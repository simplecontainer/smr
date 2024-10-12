package internal

import (
	"context"
	"github.com/docker/docker/api/types"
	dockerClient "github.com/docker/docker/client"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
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

func NewNetworks(networks []v1.ContainerNetwork) *Networks {
	networksObj := &Networks{
		Networks: make([]*Network, 0),
	}

	for _, network := range networks {
		networksObj.Add(network)
	}

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
