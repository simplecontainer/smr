package command

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
	"github.com/simplecontainer/smr/pkg/contracts/icommand"
	"github.com/spf13/cobra"
)

type Builder struct {
	parent    string
	name      string
	flags     func(cmd *cobra.Command)
	args      func(*cobra.Command, []string) error
	condition func(iapi.Api, *client.Client) bool
	command   func(iapi.Api, *client.Client, []string)
	dependsOn []func(iapi.Api, *client.Client, []string)
}

type FlagConfig struct {
	Name         string
	Shorthand    string
	DefaultValue interface{}
	Usage        string
	Required     bool
	FlagType     FlagType
}

type FlagType int

const (
	StringFlag FlagType = iota
	IntFlag
	BoolFlag
	StringSliceFlag
	IntSliceFlag
)

type CommandGroup struct {
	name     string
	commands []icommand.Command
}

func NewBuilder() *Builder {
	return &Builder{
		args:      cobra.NoArgs,
		flags:     EmptyFlag,
		condition: EmptyCondition,
		dependsOn: EmptyDepend,
		command:   EmptyFunction,
	}
}

func NewGroup(name string) *CommandGroup {
	return &CommandGroup{
		name:     name,
		commands: []icommand.Command{},
	}
}

func (cb *Builder) Parent(parent string) *Builder {
	cb.parent = parent
	return cb
}

func (cb *Builder) Name(name string) *Builder {
	cb.name = name
	return cb
}

func (cb *Builder) Flags(flags func(cmd *cobra.Command)) *Builder {
	cb.flags = flags
	return cb
}

func (cb *Builder) Args(args func(*cobra.Command, []string) error) *Builder {
	cb.args = args
	return cb
}

func (cb *Builder) NoArgs() *Builder {
	cb.args = cobra.NoArgs
	return cb
}

func (cb *Builder) ExactArgs(n int) *Builder {
	cb.args = cobra.ExactArgs(n)
	return cb
}

func (cb *Builder) MinimumNArgs(n int) *Builder {
	cb.args = cobra.MinimumNArgs(n)
	return cb
}

func (cb *Builder) Function(fn func(iapi.Api, *client.Client, []string)) *Builder {
	cb.command = fn
	return cb
}

func (cb *Builder) Condition(fn func(iapi.Api, *client.Client) bool) *Builder {
	cb.condition = fn
	return cb
}

func (cb *Builder) DependsOn(fns ...func(iapi.Api, *client.Client, []string)) *Builder {
	cb.dependsOn = append(cb.dependsOn, fns...)
	return cb
}

func (cb *Builder) Build() icommand.Command {
	return Command{
		Parent:    cb.parent,
		Name:      cb.name,
		Args:      cb.args,
		Flags:     cb.flags,
		Command:   cb.command,
		Condition: cb.condition,
		DependsOn: cb.dependsOn,
	}
}

func (cg *CommandGroup) AddRootCommand() *CommandGroup {
	cmd := NewBuilder().
		Parent("smr").
		Name(cg.name).
		Build()

	cg.commands = append(cg.commands, cmd)
	return cg
}

func (cg *CommandGroup) AddCommand(name string) *Builder {
	return NewBuilder().Parent(cg.name).Name(name)
}

func (cg *CommandGroup) AddBuiltCommand(cmd Command) *CommandGroup {
	cg.commands = append(cg.commands, cmd)
	return cg
}

func (cg *CommandGroup) GetCommands() []icommand.Command {
	return cg.commands
}

func (cb *Builder) Validate() error {
	if cb.name == "" {
		return fmt.Errorf("command name is required")
	}
	if cb.parent == "" {
		return fmt.Errorf("command parent is required")
	}
	return nil
}

func (cb *Builder) BuildWithValidation() (icommand.Command, error) {
	if err := cb.Validate(); err != nil {
		return Command{}, err
	}
	return cb.Build(), nil
}

//func ExampleAgentCommands() []Engine {
//	agentGroup := NewGroup("agent")
//
//	agentGroup.AddRootCommand()
//
//	startCmd := agentGroup.AddCommand("start").
//		RequiredStringFlag("raft", "raft endpoint (required)").
//		StringFlag("node", "simplecontainer-node-1", "Node container name").
//		Function(func(api iapi.Api, args []string) {
//			fmt.Println("Starting engine agent...")
//		}).
//		Build()
//
//	agentGroup.AddBuiltCommand(startCmd)
//
//	exportCmd := agentGroup.AddCommand("export").
//		StringFlag("api", "localhost:1443", "Public/private facing endpoint").
//		StringFlag("node", "simplecontainer-node-1", "Node name").
//		EngineFunction(func(api iapi.Api, args []string) {
//			fmt.Println("Exporting engine agent...")
//		}).
//		BuildEngine()
//
//	agentGroup.AddBuiltCommand(exportCmd)
//
//	return agentGroup.GetCommands()
//}

//func ExampleMixedCommands() ([]Engine, []Client) {
//	engineCommands := []Engine{
//		CreateSimpleEngineCommand("server", "start", func(api iapi.Api, args []string) {
//			fmt.Println("Starting server...")
//		}),
//		CreateEngineCommandWithFlags("server", "config",
//			func(api iapi.Api, args []string) {
//				fmt.Println("Configuring server...")
//			},
//			map[string]string{
//				"host": "Host address",
//				"port": "Port number",
//			},
//		),
//	}
//
//	clientCommands := []Client{
//		CreateSimpleClientCommand("client", "connect", func(client *client.Client, args []string) {
//			fmt.Println("Connecting client...")
//		}),
//		CreateClientCommandWithRequiredArgs("client", "send", 2,
//			func(client *client.Client, args []string) {
//				fmt.Printf("Sending %s to %s\n", args[0], args[1])
//			},
//		),
//	}
//
//	return engineCommands, clientCommands
//}
