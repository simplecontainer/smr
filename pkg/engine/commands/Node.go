package commands

import (
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/command"
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
	"github.com/simplecontainer/smr/pkg/engine/node"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Node() {
	Commands = append(Commands,
		command.NewBuilder().Parent("smr").Name("node").BuildWithValidation(),
		command.NewBuilder().Parent("node").Name("create").Function(cmdNodeCreate).Flags(cmdNodeCreateFlags).BuildWithValidation(),
		command.NewBuilder().Parent("node").Name("start").Function(cmdNodeStart).Flags(cmdNodeStartFlags).BuildWithValidation(),
		command.NewBuilder().Parent("node").Name("clean").Function(cmdNodeClean).Flags(cmdNodeCleanFlags).BuildWithValidation(),
		command.NewBuilder().Parent("node").Name("logs").Function(cmdNodeLogs).Flags(cmdNodeLogsFlags).BuildWithValidation(),
		command.NewBuilder().Parent("node").Name("ip").Function(cmdNodeNetworks).Flags(cmdNodeNetworksFlags).BuildWithValidation(),
	)
}

func cmdNodeCreate(api iapi.Api, cli *client.Client, args []string) {
	node.Create(api)
}
func cmdNodeCreateFlags(cmd *cobra.Command) {
	cmd.Flags().String("platform", static.PLATFORM_DOCKER, "Container platform to manage containers lifecycle")

	cmd.Flags().String("node", "simplecontainer-node", "Node container name")

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

}

func cmdNodeStart(api iapi.Api, cli *client.Client, args []string) {
	node.Start(viper.GetString("entrypoint"), viper.GetString("args"))
	node.SetupAccess()
}
func cmdNodeStartFlags(cmd *cobra.Command) {
	cmd.Flags().String("node", "simplecontainer-node", "Node container name")
	cmd.Flags().String("entrypoint", "/opt/smr/smr", "Entrypoint")
	cmd.Flags().String("args", "start", "Args")
	cmd.Flags().BoolP("y", "y", false, "Say yes to overwrite of context")
}

func cmdNodeClean(api iapi.Api, cli *client.Client, args []string) {
	node.Clean()
}
func cmdNodeCleanFlags(cmd *cobra.Command) {
	cmd.Flags().String("node", "simplecontainer-node", "Node container name")
}

func cmdNodeLogs(api iapi.Api, cli *client.Client, args []string) {
	node.Logs()
}
func cmdNodeLogsFlags(cmd *cobra.Command) {
	cmd.Flags().String("node", "simplecontainer-node", "Node container name")
}

func cmdNodeNetworks(api iapi.Api, cli *client.Client, args []string) {
	node.Networks()
}
func cmdNodeNetworksFlags(cmd *cobra.Command) {
	cmd.Flags().String("node", "simplecontainer-node", "Node")
	cmd.Flags().String("network", "bridge", "Network name")
}
