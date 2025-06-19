package commands

import (
	"github.com/simplecontainer/smr/pkg/command"
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
	"github.com/simplecontainer/smr/pkg/engine/node"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Node() {
	Commands = append(Commands,
		command.Engine{
			Parent:    "smr",
			Name:      "node",
			Condition: EmptyCondition,
			Functions: []func(iapi.Api, []string){
				func(api iapi.Api, args []string) {

				},
			},
			DependsOn: EmptyDepend,
			Flags:     EmptyFlag,
		},
		command.Engine{
			Parent:    "node",
			Name:      "start",
			Condition: EmptyCondition,
			Args:      cobra.NoArgs,
			Functions: []func(iapi.Api, []string){
				func(api iapi.Api, args []string) {
					node.Start(viper.GetString("entrypoint"), viper.GetString("args"))
					node.SetupAccess()
				},
			},
			DependsOn: EmptyDepend,
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("node", "simplecontainer-node-1", "Node container name")
				cmd.Flags().String("entrypoint", "/opt/smr/smr", "Entrypoint")
				cmd.Flags().String("args", "start", "Args")
				cmd.Flags().BoolP("y", "y", false, "Say yes to overwrite of context")
			},
		},
		command.Engine{
			Parent:    "node",
			Name:      "clean",
			Condition: EmptyCondition,
			Args:      cobra.NoArgs,
			Functions: []func(iapi.Api, []string){
				func(api iapi.Api, args []string) {
					node.Clean()
				},
			},
			DependsOn: EmptyDepend,
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("node", "simplecontainer-node-1", "Node container name")
			},
		},
		command.Engine{
			Parent:    "node",
			Name:      "create",
			Condition: EmptyCondition,
			Args:      cobra.NoArgs,
			Functions: []func(iapi.Api, []string){
				func(api iapi.Api, args []string) {
					node.Create(api)
				},
			},
			DependsOn: EmptyDepend,
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("platform", static.PLATFORM_DOCKER, "Container platform to manage containers lifecycle")

				cmd.Flags().String("node", "simplecontainer-node-1", "Node container name")

				cmd.Flags().String("image", "quay.io/simplecontainer/smr", "Node image name")
				cmd.Flags().String("tag", "latest", "Node image tag")
				cmd.Flags().String("entrypoint", "/opt/smr/smr", "Entrypoint for the smr")
				cmd.Flags().String("args", "start", "args")
				cmd.Flags().String("raft", "", "Raft Api")
				cmd.Flags().String("peer", "", "Peer for entering cluster first time. Format: https://host:port")
				cmd.Flags().Bool("join", false, "Join the raft")

				cmd.Flags().String("listen", "0.0.0.0:1443", "Simplecontainer mTLS listening interface and port combo")
				cmd.Flags().String("domain", "", "Domain that TLS certificates is valid for")
				cmd.Flags().String("ip", "", "IP address that TLS certificates is valid for")

				cmd.Flags().String("port.control", ":1443", "Port mapping of node control plane -> Default 0.0.0.0:1443")
				cmd.Flags().String("port.overlay", ":9212", "Port mapping of node overlay raft port  -> Default 0.0.0.0:9212")
				cmd.Flags().String("port.etcd", "2379", "Port mapping of node overlay raft port  -> Default 127.0.0.1:2379 (Cant be exposed to outside!)")
			},
		},
		command.Engine{
			Parent:    "node",
			Name:      "logs",
			Condition: EmptyCondition,
			Functions: []func(iapi.Api, []string){
				func(api iapi.Api, args []string) {
					node.Logs()
				},
			},
			DependsOn: EmptyDepend,
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("node", "simplecontainer-node-1", "Node")
			},
		},
		command.Engine{
			Parent:    "node",
			Name:      "networks",
			Condition: EmptyCondition,
			Functions: []func(iapi.Api, []string){
				func(api iapi.Api, args []string) {
					node.Networks()
				},
			},
			DependsOn: EmptyDepend,
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("node", "simplecontainer-node-1", "Node")
				cmd.Flags().String("network", "bridge", "Network name")
			},
		},
	)
}
