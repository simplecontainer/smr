package implementation

import (
	"context"
	"encoding/json"
	TDNetwork "github.com/docker/docker/api/types/network"
	IDClient "github.com/docker/docker/client"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
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
	cli, err := IDClient.NewClientWithOpts(IDClient.FromEnv)

	if err != nil {
		return err
	}

	newNetwork := TDNetwork.CreateOptions{IPAM: &TDNetwork.IPAM{
		Driver: "default",
		Config: []TDNetwork.IPAMConfig{TDNetwork.IPAMConfig{
			Subnet: network.IPV4AddressPool,
		}},
	}}

	_, err = cli.NetworkCreate(context.Background(), network.Name, newNetwork)

	return err
}

func (network *Network) Remove() error {
	cli, err := IDClient.NewClientWithOpts(IDClient.FromEnv)

	if err != nil {
		return err
	}

	err = cli.NetworkRemove(context.Background(), network.Name)

	return err
}

func (network *Network) Find() (map[string]TDNetwork.EndpointResource, bool, error) {
	ctx := context.Background()
	cli, err := IDClient.NewClientWithOpts(IDClient.FromEnv, IDClient.WithAPIVersionNegotiation())

	if err != nil {
		return nil, false, err
	}

	defer func(cli *IDClient.Client) {
		err = cli.Close()
		if err != nil {
			return
		}
	}(cli)

	var networks []TDNetwork.Summary
	networks, err = cli.NetworkList(ctx, TDNetwork.ListOptions{})

	if err != nil {
		return nil, false, err
	}

	for _, ntwrk := range networks {
		if ntwrk.Name == network.Name {
			return ntwrk.Containers, true, nil
		}
	}

	return nil, false, nil
}
