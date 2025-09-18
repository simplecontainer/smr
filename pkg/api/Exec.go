package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/mattn/go-shellwords"
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/exec"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/proxy/wss"
	"github.com/simplecontainer/smr/pkg/static"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"strings"
)

var wssUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (a *Api) Exec(c *gin.Context) {
	prefix := c.Param("prefix")
	version := c.Param("version")
	category := c.Param("category")
	kind := c.Param("kind")
	group := c.Param("group")
	name := c.Param("name")

	command := c.Query("command")
	height := c.Query("height")
	width := c.Query("width")

	interactive, err := strconv.ParseBool(c.Param("interactive"))
	if err != nil {
		interactive = false
	}

	format := f.New(prefix, version, category, kind, group, name)

	logger.Log.Info("exec request initiated",
		zap.String("container", format.ToString()),
		zap.String("command", command),
		zap.Bool("interactive", interactive))

	conn, err := wssUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Log.Error("failed to upgrade websocket", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upgrade websocket"})
		return
	}
	defer conn.Close()

	container := a.KindsRegistry[static.KIND_CONTAINERS].GetShared().(*shared.Shared).Registry.Find(static.SMR_PREFIX, group, name)
	if container == nil {
		logger.Log.Warn("container not found", zap.String("container", fmt.Sprintf("%s/%s", group, name)))
		sendWebSocketTextAndClose(conn, "container not found")
		return
	}

	if container.IsGhost() {
		httpClient, ok := a.Manager.Http.Clients[container.GetRuntime().Node.NodeName]
		if !ok {
			sendWebSocketTextAndClose(conn, fmt.Sprintf("node for %s '%s/%s' not found", static.KIND_CONTAINERS, group, name))
			return
		}

		url := fmt.Sprintf("%s/api/v1/exec/%s/%v", httpClient.API, format.ToString(), interactive)

		err = remoteExec(c, conn, url, httpClient)

		if err != nil && !errors.Is(err, io.EOF) {
			logger.Log.Debug("remote exec closed with error", zap.Error(err))
		} else {
			logger.Log.Debug("remote exec closed with success")
		}
	} else {
		err = localExec(c, conn, container, command, interactive, height, width)

		if err != nil && !errors.Is(err, io.EOF) {
			logger.Log.Debug("local exec session closed with error", zap.Error(err))
		} else {
			logger.Log.Debug("local exec session closed with success")
		}
	}
}

func remoteExec(c *gin.Context, conn *websocket.Conn, url string, httpClient *clients.Client) error {
	ctx, fn := context.WithCancel(c)
	proxy, err := wss.New(ctx, fn, httpClient.Http, c.Request.Header, conn, url)

	if err != nil {
		return errors.New("failed to create proxy to remote node")
	} else {
		err = proxy.WSS()

		if err != nil {
			return err
		}

		return nil
	}
}

func localExec(c *gin.Context, clientConn *websocket.Conn, container platforms.IContainer, command string, interactive bool, height string, width string) error {
	ctx, fn := context.WithCancel(c)

	parsed := strings.TrimPrefix(command, "/")
	execArgs, err := shellwords.Parse(parsed)
	if err != nil {
		panic(err)
	}

	var proxy *exec.Session
	proxy, err = exec.Create(ctx, fn, clientConn, container, execArgs, interactive, height, width)

	if err != nil {
		return err
	} else {
		err = proxy.Exec()

		if err != nil {
			return err
		}

		return nil
	}
}

func sendWebSocketTextAndClose(conn *websocket.Conn, message string) {
	closeWebSocketGracefully(conn, message)
}

func closeWebSocketGracefully(conn *websocket.Conn, reason string) {
	conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, reason))
}
