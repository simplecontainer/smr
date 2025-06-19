package commands

import (
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/command"
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
	"github.com/simplecontainer/smr/pkg/control"
	"github.com/simplecontainer/smr/pkg/control/factory"
	"github.com/simplecontainer/smr/pkg/engine/agent"
	"github.com/simplecontainer/smr/pkg/engine/controller"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Agent() {
	Commands = append(Commands,
		command.NewBuilder().Parent("smr").Name("agent").Build(),
		command.NewBuilder().Parent("agent").Name("start").Function(cmdAgentStart).Flags(cmdAgentStartFlags).Build(),
		command.NewBuilder().Parent("agent").Name("export").Function(cmdAgentExport).Flags(cmdAgentExportFlags).Build(),
		command.NewBuilder().Parent("agent").Name("import").Args(cobra.ExactArgs(2)).Function(cmdAgentImport).Flags(cmdAgentImportFlags).Build(),
		command.NewBuilder().Parent("agent").Name("stop").Function(cmdAgentStop).Build(),
		command.NewBuilder().Parent("agent").Name("control").Function(cmdAgentControl).Flags(cmdAgentControlFlags).Build(),
		command.NewBuilder().Parent("agent").Name("events").Function(cmdAgentEvents).Flags(cmdAgentControlFlags).Build(),
		command.NewBuilder().Parent("agent").Name("drain").Function(cmdAgentDrain).Flags(cmdAgentDrainFlags).Build(),
		command.NewBuilder().Parent("agent").Name("restart").Function(cmdAgentRestart).Flags(cmdAgentRestartFlags).Build(),
		command.NewBuilder().Parent("agent").Name("upgrade").Args(cobra.ExactArgs(2)).Function(cmdAgentUpgrade).Flags(cmdAgentUpgradeFlags).Build(),
	)
}

func cmdAgentStart(api iapi.Api, cli *client.Client, args []string) {
	conf := loadConfiguration()
	parsed, err := helpers.EnforceHTTPS(viper.GetString("raft"))

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	b := control.NewCommandBatch()

	b.AddCommand(factory.NewCommand("start", map[string]string{
		"raft":    parsed.String(),
		"cidr":    conf.Flannel.CIDR,
		"backend": conf.Flannel.Backend,
	}))

	agent.Start(b)
	agent.Flannel()
}
func cmdAgentStartFlags(cmdAgent *cobra.Command) {
	cmdAgent.Flags().String("raft", "", "raft endpoint")
	cmdAgent.Flags().String("node", "simplecontainer-node-1", "Node container name")
}

func cmdAgentDrain(api iapi.Api, cli *client.Client, args []string) {
	conf := loadConfiguration()
	ctrl := factory.NewCommand("drain", map[string]string{})

	b := control.NewCommandBatch()
	b.SetNodeID(conf.KVStore.Node.NodeID)
	b.AddCommand(ctrl)

	agent.Batch(b)
}
func cmdAgentDrainFlags(cmdAgent *cobra.Command) {
	cmdAgent.Flags().String("node", "simplecontainer-node-1", "Node")
	cmdAgent.Flags().String("wait", "", "Node")
}

func cmdAgentRestart(api iapi.Api, cli *client.Client, args []string) {
	conf := loadConfiguration()
	ctrl := factory.NewCommand("drain", map[string]string{})
	restart := factory.NewCommand("restart", map[string]string{})

	b := control.NewCommandBatch()
	b.SetNodeID(conf.KVStore.Node.NodeID)
	b.AddCommand(ctrl)
	b.AddCommand(restart)

	agent.Batch(b)
}
func cmdAgentRestartFlags(cmdAgent *cobra.Command) {
	cmdAgent.Flags().String("node", "simplecontainer-node-1", "Node")
	cmdAgent.Flags().String("wait", "", "Node")
}

func cmdAgentUpgrade(api iapi.Api, cli *client.Client, args []string) {
	conf := loadConfiguration()
	ctrl := factory.NewCommand("drain", map[string]string{})
	upgrade := factory.NewCommand("upgrade", map[string]string{"image": args[0], "tag": args[1]})

	b := control.NewCommandBatch()
	b.SetNodeID(conf.KVStore.Node.NodeID)
	b.AddCommand(ctrl)
	b.AddCommand(upgrade)

	agent.Batch(b)
}
func cmdAgentUpgradeFlags(cmdAgent *cobra.Command) {
	cmdAgent.Flags().String("node", "simplecontainer-node-1", "Node")
	cmdAgent.Flags().String("wait", "", "Node")
}

func cmdAgentExport(api iapi.Api, cli *client.Client, args []string) {
	agent.Export()
}
func cmdAgentExportFlags(cmdAgent *cobra.Command) {
	cmdAgent.Flags().String("api", "localhost:1443", "Public/private facing endpoint for control plane. eg example.com:1443")
	cmdAgent.Flags().String("node", "simplecontainer-node-1", "Node")
}

func cmdAgentImport(api iapi.Api, cli *client.Client, args []string) {
	agent.Import(args[0], args[1])
}
func cmdAgentImportFlags(cmdAgent *cobra.Command) {
	cmdAgent.Flags().String("node", "simplecontainer-node-1", "Node")
	cmdAgent.Flags().BoolP("y", "y", false, "Say yes to overwrite of context")
}

func cmdAgentStop(api iapi.Api, cli *client.Client, args []string) {
	agent.Stop()
}

func cmdAgentControl(api iapi.Api, cli *client.Client, args []string) {
	controller.Control(api)
}
func cmdAgentControlFlags(cmdAgent *cobra.Command) {
	cmdAgent.Flags().String("node", "simplecontainer-node-1", "Node")
	cmdAgent.Flags().BoolP("y", "y", false, "Say yes to overwrite of context")
}

func cmdAgentEvents(api iapi.Api, cli *client.Client, args []string) {
	agent.Events()
}
func cmdAgentEventsFlags(cmdAgent *cobra.Command) {
	cmdAgent.Flags().String("node", "simplecontainer-node-1", "Node")
	cmdAgent.Flags().String("wait", "", "Node")
}
