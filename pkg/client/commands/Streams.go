package commands

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/client/exec"
	"github.com/simplecontainer/smr/pkg/command"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/wss"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/term"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func Streams() {
	Commands = append(Commands,
		command.Client{
			Parent: "smrctl",
			Name:   "debug",
			Condition: func(*client.Client) bool {
				return true
			},
			Args: cobra.ExactArgs(1),
			Functions: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {
					format, err := helpers.BuildFormat(args[0], cli.Group)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					resp, err := network.Raw(cli.Context.GetClient(), fmt.Sprintf("%s/api/v1/debug/%s/%s/%s", cli.Context.APIURL, format.ToString(), viper.GetString("container"), strconv.FormatBool(viper.GetBool("f"))), http.MethodGet, nil)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					sigChan := make(chan os.Signal, 1)
					signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
					go func() {
						<-sigChan
						resp.Body.Close()
						cancel()
					}()

					err = helpers.PrintBytes(ctx, resp.Body)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}
				},
			},
			DependsOn: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {},
			},
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("container", "main", "Logs from main or init")
				cmd.Flags().Bool("f", false, "Follow logs")
			},
		},
		command.Client{
			Parent: "smrctl",
			Name:   "logs",
			Condition: func(*client.Client) bool {
				return true
			},
			Args: cobra.NoArgs,
			Functions: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {
					format, err := helpers.BuildFormat(args[0], cli.Group)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					resp, err := network.Raw(cli.Context.GetClient(), fmt.Sprintf("%s/api/v1/logs/%s/%s/%s", cli.Context.APIURL, format.ToString(), viper.GetString("container"), strconv.FormatBool(viper.GetBool("f"))), http.MethodGet, nil)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					sigChan := make(chan os.Signal, 1)
					signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
					go func() {
						<-sigChan
						resp.Body.Close()
						cancel()
					}()

					if resp.StatusCode == 200 {
						err = helpers.PrintBytesDemux(ctx, resp.Body)

						if err != nil {
							helpers.PrintAndExit(err, 1)
						}
					} else {
						err = helpers.PrintBytes(ctx, resp.Body)

						if err != nil {
							helpers.PrintAndExit(err, 1)
						}
					}
				},
			},
			DependsOn: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {},
			},
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("container", "main", "Logs from main or init")
				cmd.Flags().Bool("f", false, "Follow logs")
			},
		},
		command.Client{
			Parent: "smrctl",
			Name:   "exec",
			Condition: func(*client.Client) bool {
				return true
			},
			Args: cobra.ExactArgs(1),
			Functions: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {
					format, err := helpers.BuildFormat(args[0], cli.Group)
					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					interactive := viper.GetBool("it")
					command := viper.GetString("c")

					url := fmt.Sprintf("%s/api/v1/exec/%s/%s", cli.Context.APIURL, format.ToString(), strconv.FormatBool(interactive))

					requestHeaders := http.Header{}
					requestHeaders.Add("command", command)

					conn, err := wss.Request(cli.Context.GetClient(), requestHeaders, url)
					if err != nil {
						helpers.PrintAndExit(err, 1)
					}
					defer conn.Close()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					sigChan := make(chan os.Signal, 1)
					signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
					go func() {
						<-sigChan
						logger.Log.Debug("received termination signal")
						cancel()
					}()

					handle(ctx, cancel, conn, interactive)
				},
			},
			DependsOn: []func(*client.Client, []string){
				func(cli *client.Client, args []string) {},
			},
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("c", "/bin/sh", "Command to execute in container")
				cmd.Flags().Bool("it", false, "Interactive session")
			},
		},
	)
}

func handle(ctx context.Context, cancel context.CancelFunc, conn *websocket.Conn, interactive bool) {
	server := make(chan error)
	client := make(chan error)

	if interactive {
		fd := int(os.Stdin.Fd())
		isTerminal := term.IsTerminal(fd)

		var oldState *term.State
		var err error

		if isTerminal {
			oldState, err = term.MakeRaw(fd)

			if err != nil {
				logger.Log.Error("error setting terminal to raw mode", zap.Error(err))
				os.Exit(1)
			}

			defer term.Restore(fd, oldState)
		}

		go func() {
			err := exec.Write(ctx, conn)

			select {
			case client <- err:
			}
		}()
	}

	go func() {
		err := exec.Read(ctx, conn)

		select {
		case server <- err:
		}
	}()

	select {
	case err := <-client:
		conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "client terminated connection"))

		if err != nil {
			fmt.Println(err)
		}

		return
	case err := <-server:
		if err != nil {
			fmt.Println(err)
		}

		return
	case <-ctx.Done():
		conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "client terminated connection"))
		return
	}
}
