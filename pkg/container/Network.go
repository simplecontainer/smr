package container

import (
	"context"
	"errors"
	"fmt"
	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/qdnqn/smr/pkg/logger"
	"net"
)

const STARTING_SUBNET string = "10.10.0.0/16"

func (container *Container) CreateNetwork() error {
	if !container.FindNetwork() {
		cli, err := client.NewClientWithOpts(client.FromEnv)
		if err != nil {
			logger.Log.Fatal(err.Error())
		}

		for _, netw := range container.Static.Networks {
			subnet, err := container.GenerateNetworkSubnet()

			if err != nil {
				return err
			}

			dnetw := types.NetworkCreate{IPAM: &network.IPAM{
				Driver: "default",
				Config: []network.IPAMConfig{network.IPAMConfig{
					Subnet: subnet,
				}},
			}}

			res, err := cli.NetworkCreate(context.Background(), netw, dnetw)
			if err != nil {
				return err
			}

			container.Runtime.Networks[netw] = Network{
				NetworkId: res.ID,
				IP:        "",
			}
		}
	}

	return nil
}

func (container *Container) FindNetwork() bool {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	networks, err := cli.NetworkList(ctx, types.NetworkListOptions{})

	if err != nil {
		panic(err)
	}

	for _, network := range networks {
		for _, netw := range container.Static.Networks {
			if network.Name == netw {
				return true
			}
		}
	}

	return false
}

func (container *Container) GetNetwork() *network.NetworkingConfig {
	dnetw := network.NetworkingConfig{EndpointsConfig: map[string]*network.EndpointSettings{}}

	if container.Static.NetworkMode != "host" {
		for _, netw := range container.Static.Networks {
			dnetw.EndpointsConfig[netw] = &network.EndpointSettings{
				NetworkID: netw,
			}
		}
	}

	return &dnetw
}

func (container *Container) GenerateNetworkSubnet() (string, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	_, IPv4net, err := net.ParseCIDR(STARTING_SUBNET)

	logger.Log.Info(fmt.Sprintf("generating network IPV4: %s", IPv4net))

	networks, err := cli.NetworkList(ctx, types.NetworkListOptions{})
	subnets := make([]*net.IPNet, 0)

	if err != nil {
		return "", err
	}

	for _, network := range networks {
		if len(network.IPAM.Config) > 0 {
			if network.IPAM.Config[0].Subnet != "" {
				_, tempIPv4net, err := net.ParseCIDR(network.IPAM.Config[0].Subnet)

				if err != nil {
					return "", errors.New("invalid CIDR detected")
				}

				subnets = append(subnets, tempIPv4net)
			}
		}
	}

	for _, subnet := range subnets {
		for intersect(IPv4net, subnet) {
			IPv4net, _ = cidr.NextSubnet(IPv4net, 32)
		}
	}

	return IPv4net.String(), nil
}

func (container *Container) FindNetworkAlias(endpointName string, networkId string) bool {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
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

func intersect(n1, n2 *net.IPNet) bool {
	return n2.Contains(n1.IP) || n1.Contains(n2.IP)
}
