package container

import (
	"context"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"smr/pkg/logger"
	"smr/pkg/static"
)

func (container *Container) AgentConnectToTheSameNetwork(containerId string, networkId string) bool {
	if c := container.Get(); c != nil && c.State == "running" {
		ctx := context.Background()
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}
		defer cli.Close()

		smrAgentEndpointSettings := container.GenerateAgentNetwork(networkId)

		if !container.FindNetworkAlias(static.SMR_ENDPOINT_NAME, networkId) {
			err = cli.NetworkConnect(ctx, networkId, containerId, smrAgentEndpointSettings)

			if err != nil {
				logger.Log.Error(err.Error())
				return false
			}
		}

		return true
	} else {
		return false
	}
}

func (container *Container) GenerateAgentNetwork(networkId string) *network.EndpointSettings {
	return &network.EndpointSettings{
		NetworkID: networkId,
	}
}
