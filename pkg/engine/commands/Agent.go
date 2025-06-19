package commands

import (
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/command"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
	"github.com/simplecontainer/smr/pkg/control"
	"github.com/simplecontainer/smr/pkg/control/factory"
	"github.com/simplecontainer/smr/pkg/engine/agent"
	"github.com/simplecontainer/smr/pkg/engine/controller"
	"github.com/simplecontainer/smr/pkg/startup"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Agent() {
	Commands = append(Commands,
		command.Engine{
			Parent:    "smr",
			Name:      "agent",
			Condition: EmptyCondition,
			Args:      cobra.NoArgs,
			Functions: EmptyDepend,
			DependsOn: EmptyDepend,
			Flags:     EmptyFlag,
		},
		command.Engine{
			Parent:    "agent",
			Name:      "start",
			Condition: EmptyCondition,
			Args:      cobra.NoArgs,
			Functions: []func(iapi.Api, []string){
				func(api iapi.Api, args []string) {
					environment := configuration.NewEnvironment(configuration.WithHostConfig())
					conf, err := startup.Load(environment)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

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
				},
			},
			DependsOn: EmptyDepend,
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("raft", "", "raft endpoint")
				cmd.Flags().String("node", "simplecontainer-node-1", "Node container name")
			},
		},
		command.Engine{
			Parent:    "agent",
			Name:      "export",
			Condition: EmptyCondition,
			Args:      cobra.NoArgs,
			Functions: []func(iapi.Api, []string){
				func(api iapi.Api, args []string) {
					agent.Export()
				},
			},
			DependsOn: EmptyDepend,
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("api", "localhost:1443", "Public/private facing endpoint for control plane. eg example.com:1443")
				cmd.Flags().String("node", "simplecontainer-node-1", "Node")
			},
		},
		command.Engine{
			Parent:    "agent",
			Name:      "import",
			Condition: EmptyCondition,
			Args:      cobra.ExactArgs(2),
			Functions: []func(iapi.Api, []string){
				func(api iapi.Api, args []string) {
					agent.Import(args[0], args[1])
				},
			},
			DependsOn: EmptyDepend,
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("node", "simplecontainer-node-1", "Node")
				cmd.Flags().BoolP("y", "y", false, "Say yes to overwrite of context")
			},
		},
		command.Engine{
			Parent:    "agent",
			Name:      "stop",
			Condition: EmptyCondition,
			Functions: []func(iapi.Api, []string){
				func(api iapi.Api, args []string) {
					agent.Stop()
				},
			},
			DependsOn: EmptyDepend,
			Flags:     EmptyFlag,
		},
		command.Engine{
			Parent:    "agent",
			Name:      "control",
			Condition: EmptyCondition,
			Functions: []func(iapi.Api, []string){
				func(api iapi.Api, args []string) {
					controller.Control(api)
				},
			},
			DependsOn: EmptyDepend,
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("node", "simplecontainer-node-1", "Node")
			},
		},
		command.Engine{
			Parent:    "agent",
			Name:      "events",
			Condition: EmptyCondition,
			Functions: []func(iapi.Api, []string){
				func(api iapi.Api, args []string) {
					agent.Events()
				},
			},
			DependsOn: EmptyDepend,
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("node", "simplecontainer-node-1", "Node")
				cmd.Flags().String("wait", "", "Node")
			},
		},
		command.Engine{
			Parent:    "agent",
			Name:      "drain",
			Condition: EmptyCondition,
			Functions: []func(iapi.Api, []string){
				func(api iapi.Api, args []string) {
					environment := configuration.NewEnvironment(configuration.WithHostConfig())
					conf, err := startup.Load(environment)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					ctrl := factory.NewCommand("drain", map[string]string{})

					b := control.NewCommandBatch()
					b.SetNodeID(conf.KVStore.Node.NodeID)
					b.AddCommand(ctrl)

					agent.Batch(b)
				},
			},
			DependsOn: EmptyDepend,
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("node", "simplecontainer-node-1", "Node")
				cmd.Flags().String("wait", "", "Node")
			},
		},
		command.Engine{
			Parent:    "agent",
			Name:      "restart",
			Condition: EmptyCondition,
			Args:      cobra.NoArgs,
			Functions: []func(iapi.Api, []string){
				func(api iapi.Api, args []string) {
					environment := configuration.NewEnvironment(configuration.WithHostConfig())
					conf, err := startup.Load(environment)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					ctrl := factory.NewCommand("drain", map[string]string{})
					restrt := factory.NewCommand("restart", map[string]string{})

					b := control.NewCommandBatch()
					b.SetNodeID(conf.KVStore.Node.NodeID)
					b.AddCommand(ctrl)
					b.AddCommand(restrt)

					agent.Batch(b)
				},
			},
			DependsOn: EmptyDepend,
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("node", "simplecontainer-node-1", "Node")
				cmd.Flags().String("wait", "", "Node")
			},
		},
		command.Engine{
			Parent:    "agent",
			Name:      "upgrade",
			Condition: EmptyCondition,
			Args:      cobra.ExactArgs(2),
			Functions: []func(iapi.Api, []string){
				func(api iapi.Api, args []string) {
					environment := configuration.NewEnvironment(configuration.WithHostConfig())
					conf, err := startup.Load(environment)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					ctrl := factory.NewCommand("drain", map[string]string{})
					restrt := factory.NewCommand("upgrade", map[string]string{"image": args[0], "tag": args[1]})

					b := control.NewCommandBatch()
					b.SetNodeID(conf.KVStore.Node.NodeID)
					b.AddCommand(ctrl)
					b.AddCommand(restrt)

					agent.Batch(b)
				},
			},
			DependsOn: EmptyDepend,
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("node", "simplecontainer-node-1", "Node")
				cmd.Flags().String("wait", "", "Node")
			},
		},
	)
}
