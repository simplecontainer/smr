package api

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/network/wss"
	"github.com/simplecontainer/smr/pkg/static"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var wssUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (api *Api) Exec(c *gin.Context) {
	prefix := c.Param("prefix")
	version := c.Param("version")
	category := c.Param("kind")
	kind := c.Param("containers")
	group := c.Param("group")
	name := c.Param("name")
	command := c.Param("command")

	interactive, err := strconv.ParseBool(c.Param("interactive"))
	if err != nil {
		interactive = false
	}

	format := f.New(prefix, version, category, kind, group, name)

	logger.Log.Info("Exec request",
		zap.String("container", format.ToString()),
		zap.String("command", command),
		zap.Bool("interactive", interactive))

	conn, err := wssUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Log.Error("Failed to upgrade WebSocket", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade WebSocket"})
		return
	}
	defer conn.Close()

	container := api.KindsRegistry[static.KIND_CONTAINERS].GetShared().(*shared.Shared).Registry.Find(static.SMR_PREFIX, group, name)
	if container == nil {
		logger.Log.Warn("Container not found", zap.String("container", fmt.Sprintf("%s/%s", group, name)))
		sendWebSocketTextAndClose(conn, "container not found")
		return
	}

	if container.IsGhost() {
		handleGhostContainerExec(api, conn, container, format, interactive, command)
	} else {
		if interactive {
			handleInteractiveExec(conn, container, command)
		} else {
			handleNonInteractiveExec(conn, container, command)
		}
	}
}

// TODO: Group and refactor

func handleGhostContainerExec(api *Api, conn *websocket.Conn, container platforms.IContainer, format f.Format, interactive bool, command string) {
	runtime := container.GetRuntime()
	nodeName := runtime.Node.NodeName

	client, ok := api.Manager.Http.Clients[nodeName]
	if !ok {
		msg := fmt.Sprintf("node client for %s not found", nodeName)
		logger.Log.Error(msg)
		sendWebSocketTextAndClose(conn, msg)
		return
	}

	url := fmt.Sprintf("%s/api/v1/exec/%s/%v%s", client.API, format.ToString(), interactive, command)
	logger.Log.Info("Proxying exec to remote node",
		zap.String("url", url),
		zap.String("node", nodeName))

	err := wss.StreamRemote(client.Http, conn, url)
	if err != nil {
		logger.Log.Error("Error proxying to remote node", zap.Error(err))
		return
	}
}

func handleNonInteractiveExec(conn *websocket.Conn, container platforms.IContainer, command string) {
	result, err := container.Exec(strings.Split(command, " "))
	if err != nil {
		logger.Log.Error("Command execution failed", zap.Error(err))
		sendWebSocketError(conn, err)
		return
	}

	bytes, err := json.Marshal(result)
	if err != nil {
		logger.Log.Error("Failed to marshal result", zap.Error(err))
		sendWebSocketError(conn, err)
		return
	}

	err = conn.WriteMessage(websocket.BinaryMessage, bytes)
	if err != nil {
		logger.Log.Error("Failed to write result", zap.Error(err))
	}

	closeWebSocketGracefully(conn, "")
}

func handleInteractiveExec(conn *websocket.Conn, container platforms.IContainer, command string) {
	execID, reader, execConn, err := container.ExecTTY(strings.Split(command, " "), true)
	if err != nil {
		logger.Log.Error("Failed to create interactive session", zap.Error(err))
		sendWebSocketError(conn, err)
		return
	}
	defer execConn.Close()

	logger.Log.Debug("Interactive session started", zap.String("execID", execID))

	var wg sync.WaitGroup
	wg.Add(2)

	done := make(chan error, 2)
	stopChan := make(chan struct{})

	go func() {
		defer wg.Done()

		buf := make([]byte, 1024)
		for {
			select {
			case <-stopChan:
				return
			default:
				n, err := reader.Read(buf)
				if err != nil {
					if err != io.EOF {
						logger.Log.Error("Error reading from container", zap.Error(err))
					}

					done <- err
					return
				}

				if n > 0 {
					err = conn.WriteMessage(websocket.BinaryMessage, buf[:n])
					if err != nil {
						logger.Log.Error("Error writing to WebSocket", zap.Error(err))
						done <- err
						return
					}
				}
			}
		}
	}()

	go func() {
		defer wg.Done()

		for {
			select {
			case <-stopChan:
				return
			default:
				messageType, msg, err := conn.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
						logger.Log.Error("Unexpected WebSocket close", zap.Error(err))
					} else {
						logger.Log.Debug("WebSocket closed by client")
					}
					done <- err
					return
				}

				if messageType == websocket.TextMessage || messageType == websocket.BinaryMessage {
					_, err = execConn.Write(msg)
					if err != nil {
						logger.Log.Error("Error writing to container", zap.Error(err))
						done <- err
						return
					}
				}
			}
		}
	}()

	timeout := time.After(60 * time.Minute)

	select {
	case err := <-done:
		close(stopChan)

		if err == io.EOF {
			exitCode, inspectErr := container.ExecInspect(execID)

			if inspectErr != nil {
				logger.Log.Error("Failed to get exit code", zap.Error(inspectErr))
				closeWebSocketGracefully(conn, "Session closed, failed to get exit code")
			} else {
				logger.Log.Info("Command completed", zap.Int("exitCode", exitCode))
				closeWebSocketGracefully(conn, fmt.Sprintf("Session closed, exit code: %d", exitCode))
			}
		} else if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
			logger.Log.Info("WebSocket closed by client")
		} else {
			logger.Log.Error("Interactive session error", zap.Error(err))
			closeWebSocketGracefully(conn, fmt.Sprintf("Error: %v", err))
		}

		wg.Wait()

	case <-timeout:
		logger.Log.Warn("Interactive session timed out")
		close(stopChan)
		wg.Wait()
		closeWebSocketGracefully(conn, "Session timed out")
	}
}

func sendWebSocketTextAndClose(conn *websocket.Conn, message string) {
	closeWebSocketGracefully(conn, message)
}

func sendWebSocketError(conn *websocket.Conn, err error) {
	closeWebSocketGracefully(conn, err.Error())
}

func closeWebSocketGracefully(conn *websocket.Conn, reason string) {
	conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, reason))
}

func parseCloseMessage(msg []byte) (int, string) {
	if len(msg) < 2 {
		return -1, ""
	}
	code := int(binary.BigEndian.Uint16(msg[:2]))
	reason := string(msg[2:])
	return code, reason
}
