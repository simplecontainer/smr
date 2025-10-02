package client

import (
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/contexts"
	"github.com/simplecontainer/smr/pkg/packer/ocicredentials"
	"github.com/simplecontainer/smr/pkg/packer/signature"
	"github.com/spf13/viper"
	"log"
)

func New(config *configuration.Configuration, rootDir string) *Client {
	cfg := contexts.DefaultConfig(rootDir)

	manager, err := contexts.NewManager(cfg)
	if err != nil {
		log.Fatalf("failed to create client manager: %v", err)
	}

	client := &Client{
		Manager: manager,
		Group:   viper.GetString("g"),
	}

	if config != nil {
		client.Credentials = ocicredentials.NewManager(config.Environment.Host.Home)
		client.Signer = signature.NewSigner()
	}

	return client
}
