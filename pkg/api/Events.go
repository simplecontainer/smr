package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/simplecontainer/smr/pkg/contracts/ievents"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/wss"
	"go.uber.org/zap"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (api *Api) Events(c *gin.Context) {
	w, r := c.Writer, c.Request

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Log.Error("Failed to upgrade WebSocket connection: ", zap.Error(err))
		return
	}
	defer conn.Close()

	api.Wss.Lock.Lock()
	ch := make(chan ievents.Event, 10) // Use buffered channel to avoid blocking
	position := len(api.Wss.Channels)
	api.Wss.Channels = append(api.Wss.Channels, ch)
	api.Wss.Lock.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go ListenEvents(ctx, api.Wss, position, conn)

	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err = conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Log.Warn("WebSocket closed unexpectedly: ", zap.Error(err))
			}
			break
		}
	}
}

func ListenEvents(ctx context.Context, wss *wss.WebSockets, position int, conn *websocket.Conn) {
	defer func() {
		wss.Lock.Lock()
		defer wss.Lock.Unlock()

		if position >= 0 && position < len(wss.Channels) {
			close(wss.Channels[position])
			wss.Channels = append(wss.Channels[:position], wss.Channels[position+1:]...)
		}
	}()

	for {
		select {
		case data := <-wss.Channels[position]:
			bytes, err := data.ToJson()
			if err != nil {
				logger.Log.Error("Failed to serialize event: ", zap.Error(err))
				continue
			}

			message, err := websocket.NewPreparedMessage(websocket.TextMessage, bytes)
			if err != nil {
				logger.Log.Error("Failed to prepare WebSocket message: ", zap.Error(err))
				continue
			}

			err = conn.WritePreparedMessage(message)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logger.Log.Warn("WebSocket write error: ", zap.Error(err))
				}
				return
			}

		case <-ctx.Done():
			return
		}
	}
}
