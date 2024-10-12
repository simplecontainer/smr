package network

import (
	"context"
	"encoding/json"
	"github.com/docker/docker/api/types"
	dockerNetwork "github.com/docker/docker/api/types/network"
	dockerClient "github.com/docker/docker/client"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/logger"
)

func New(bytes []byte) *Network {
	network := v1.NetworkDefinition{}

	err := json.Unmarshal(bytes, &network)

	if err != nil {
		return &Network{}
	}

	return &Network{
		Name:            network.Meta.Name,
		Driver:          network.Spec.Driver,
		IPV4AddressPool: network.Spec.IPV4AddressPool,
	}
}

func (network *Network) Create() error {
	found, err := network.Find()

	if err != nil {
		return err
	}

	if !found {
		var cli *dockerClient.Client
		cli, err = dockerClient.NewClientWithOpts(dockerClient.FromEnv)

		if err != nil {
			logger.Log.Fatal(err.Error())
		}

		newNetwork := types.NetworkCreate{IPAM: &dockerNetwork.IPAM{
			Driver: "default",
			Config: []dockerNetwork.IPAMConfig{dockerNetwork.IPAMConfig{
				Subnet: network.IPV4AddressPool,
			}},
		}}

		_, err = cli.NetworkCreate(context.Background(), network.Name, newNetwork)

		if err != nil {
			return err
		}
	}

	return nil
}

func (network *Network) Find() (bool, error) {
	ctx := context.Background()
	cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())

	if err != nil {
		return false, err
	}

	defer func(cli *dockerClient.Client) {
		err = cli.Close()
		if err != nil {
			return
		}
	}(cli)

	var networks []types.NetworkResource
	networks, err = cli.NetworkList(ctx, types.NetworkListOptions{})

	if err != nil {
		panic(err)
	}

	for _, ntwrk := range networks {
		if ntwrk.Name == network.Name {
			return true, nil
		}
	}

	return false, nil
}
