package command

import (
	"github.com/simplecontainer/smr/pkg/api"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/spf13/cobra"
)

type Engine struct {
	Parent    string
	Name      string
	Flag      string
	Condition func(*api.Api) bool
	Functions []func(*api.Api, []string)
	DependsOn []func(*api.Api, []string)
	Flags     func(command *cobra.Command)
}

type Client struct {
	Parent    string
	Name      string
	Flag      string
	Condition func(client *client.Client) bool
	Functions []func(*client.Client, []string)
	DependsOn []func(*client.Client, []string)
}
