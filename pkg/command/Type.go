package command

import (
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
	"github.com/spf13/cobra"
)

type Engine struct {
	Parent    string
	Name      string
	Flag      string
	Args      func(*cobra.Command, []string) error
	Condition func(iapi.Api) bool
	Functions []func(iapi.Api, []string)
	DependsOn []func(iapi.Api, []string)
	Flags     func(command *cobra.Command)
}

type Client struct {
	Parent    string
	Name      string
	Flag      string
	Args      func(*cobra.Command, []string) error
	Condition func(client *client.Client) bool
	Functions []func(*client.Client, []string)
	DependsOn []func(*client.Client, []string)
	Flags     func(command *cobra.Command)
}
