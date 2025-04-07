package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/network/wss"
	"github.com/simplecontainer/smr/pkg/static"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func (api *Api) Exec(c *gin.Context) {
	prefix := c.Param("prefix")
	version := c.Param("version")
	category := c.Param("category")
	kind := c.Param("kind")
	group := c.Param("group")
	name := c.Param("name")
	interactive, err := strconv.ParseBool(c.Param("interactive"))

	if err != nil {
		interactive = false
	}

	command := c.Param("command")

	format := f.New(prefix, version, category, kind, group, name)

	conn, err := wssUpgrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade WebSocket"})
		return
	}

	defer conn.Close()

	container := api.KindsRegistry[static.KIND_CONTAINERS].GetShared().(*shared.Shared).Registry.Find(static.SMR_PREFIX, group, name)

	if container == nil {
		conn.WriteMessage(websocket.TextMessage, []byte("container not found"))
		return
	}

	if container.IsGhost() {
		client, ok := api.Manager.Http.Clients[container.GetRuntime().Node.NodeName]

		if ok {
			wss.StreamRemote(client.Http, conn, fmt.Sprintf("https://%s/api/v1/exec/%s/%s/%s", client.API, format.ToString(), c.Param("interactive"), command))
		} else {
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(0, err.Error()))
		}
	} else {
		if !interactive {
			result, err := container.Exec(strings.Split(command, " "))

			if err != nil {
				conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(0, err.Error()))
			}

			bytes, err := json.Marshal(result)

			if err != nil {
				conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(0, err.Error()))
			}

			conn.WriteMessage(websocket.BinaryMessage, bytes)
			conn.Close()
		} else {
			execID, reader, execConn, err := container.ExecTTY(strings.Split(command, " "), interactive)

			if err != nil {
				conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(1000, err.Error()))
				return
			}

			defer execConn.Close()

			done := make(chan error)

			go func() {
				for {
					buf := make([]byte, 1024)
					n, err := reader.Read(buf)

					if err != nil && err != io.EOF {
						err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(1000, err.Error()))

						if err != nil {
							logger.Log.Error("", zap.Error(err))
						}

						return
					}

					conn.WriteMessage(websocket.BinaryMessage, buf[:n])

					if err == io.EOF {
						done <- err
						return
					}
				}
			}()

			go func() {
				for {
					_, msg, err := conn.ReadMessage()

					if err != nil {
						return
					}

					_, err = execConn.Write(msg)

					if err != nil {
						err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(1000, err.Error()))

						if err != nil {
							logger.Log.Error("", zap.Error(err))
						}

						return
					}
				}
			}()

			select {
			case err := <-done:
				if err == io.EOF {
					exitCode, err := container.ExecInspect(execID)

					if err != nil {
						conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(1000, err.Error()))
					} else {
						conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(1000, fmt.Sprintf("session closed, exit code: %d", exitCode)))
					}
				}
				break
			}
		}
	}
}
