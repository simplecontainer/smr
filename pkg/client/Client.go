package client

import "github.com/simplecontainer/smr/pkg/configuration"

func New(config *configuration.Configuration) *Client {
	return &Client{
		Context: NewContext(config.Environment.Host.Home),
	}
}
