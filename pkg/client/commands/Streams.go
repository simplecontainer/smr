package commands

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/client/exec"
	"github.com/simplecontainer/smr/pkg/command"
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
	CExec "github.com/simplecontainer/smr/pkg/exec"
	"github.com/simplecontainer/smr/pkg/f"
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
		command.NewBuilder().Parent("smrctl").Name("debug").Args(cobra.ExactArgs(1)).Function(cmdDebug).Flags(cmdDebugFlags).BuildWithValidation(),
		command.NewBuilder().Parent("smrctl").Name("logs").Args(cobra.ExactArgs(1)).Function(cmdLogs).Flags(cmdLogsFlags).BuildWithValidation(),
		command.NewBuilder().Parent("smrctl").Name("exec").Args(cobra.ExactArgs(1)).Function(cmdExec).Flags(cmdExecFlags).BuildWithValidation(),
	)
}

func cmdDebug(api iapi.Api, cli *client.Client, args []string) {
	format, err := f.Build(args[0], cli.Group)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resp, err := network.Raw(ctx, cli.Context.GetHTTPClient(), fmt.Sprintf("%s/api/v1/debug/%s/%s/%s", cli.Context.APIURL, format.ToString(), viper.GetString("container"), strconv.FormatBool(viper.GetBool("follow"))), http.MethodGet, nil)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

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
}
func cmdDebugFlags(cmd *cobra.Command) {
	cmd.Flags().String("container", "main", "Logs from main or init")
	cmd.Flags().BoolP("follow", "f", false, "Follow logs")
}

func cmdLogs(api iapi.Api, cli *client.Client, args []string) {
	format, err := f.Build(args[0], cli.Group)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resp, err := network.Raw(ctx, cli.Context.GetHTTPClient(), fmt.Sprintf("%s/api/v1/logs/%s/%s/%s", cli.Context.APIURL, format.ToString(), viper.GetString("container"), strconv.FormatBool(viper.GetBool("follow"))), http.MethodGet, nil)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		<-sigChan
		cancel()
		resp.Body.Close()
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
}
func cmdLogsFlags(cmd *cobra.Command) {
	cmd.Flags().String("container", "main", "Logs from main or init")
	cmd.Flags().BoolP("follow", "f", false, "Follow logs")
}

func cmdExec(api iapi.Api, cli *client.Client, args []string) {
	format, err := f.Build(args[0], cli.Group)
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	interactive := viper.GetBool("it")
	command := viper.GetString("c")

	width, height, err := term.GetSize(int(os.Stdin.Fd()))

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	url := fmt.Sprintf("%s/api/v1/exec/%s/%s?command=%s&width=%d&height=%d", cli.Context.APIURL, format.ToString(), strconv.FormatBool(interactive), command, width, height)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, cancelWSS, err := wss.Request(ctx, cli.Context.GetHTTPClient(), nil, url)
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}
	defer conn.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		err := cancelWSS()

		if err != nil {
			fmt.Println(err)
		}
	}()

	handle(ctx, cancel, conn, interactive)
}
func cmdExecFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("c", "c", "/bin/sh", "Command to execute in container")
	cmd.Flags().BoolP("it", "i", false, "Interactive session")
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

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGWINCH)
			go func() {
				for range sigCh {
					width, height, err := term.GetSize(fd)
					if err != nil {
						logger.Log.Warn("failed to get terminal size", zap.Error(err))
						continue
					}

					logger.Log.Debug("terminal resized", zap.Int("cols", width), zap.Int("rows", height))
					resize, err := CExec.NewResize(width, height)
					if err != nil {
						logger.Log.Warn("failed to set terminal size", zap.Error(err))
						continue
					}

					bytes, err := resize.Marshal()
					if err != nil {
						logger.Log.Warn("failed to set terminal size", zap.Error(err))
						continue
					}

					err = conn.WriteMessage(websocket.TextMessage, bytes)
					if err != nil {
						logger.Log.Warn("failed to set terminal size", zap.Error(err))
						continue
					}
				}
			}()

			syscall.Kill(syscall.Getpid(), syscall.SIGWINCH)
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
