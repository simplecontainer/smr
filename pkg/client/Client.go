package client

import (
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/spf13/viper"
)

func New(config *configuration.Configuration) *Client {
	cfg := DefaultConfig(config.Environment.Host.Home)

	return &Client{
		Context: NewContext(cfg),
		Group:   viper.GetString("g"),
	}
}
