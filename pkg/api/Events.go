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
	"sync"
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

	// Locking to safely add the new channel to the map
	api.Wss.Lock.Lock()
	ch := make(chan ievents.Event, 100) // Increased buffer size to avoid blocking
	position := len(api.Wss.Channels)   // Use position as the map key
	api.Wss.Channels[position] = ch
	api.Wss.Lock.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var closeOnce sync.Once

	go func() {
		defer func() {
			api.Wss.Lock.Lock()
			defer api.Wss.Lock.Unlock()

			closeOnce.Do(func() {
				close(ch)
			})

			// Safely remove the channel from the map
			delete(api.Wss.Channels, position)
		}()

		ListenEvents(ctx, api.Wss, position, conn)
	}()

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

	cancel()
	closeOnce.Do(func() {
		close(ch)
	})
}

func ListenEvents(ctx context.Context, wss *wss.WebSockets, position int, conn *websocket.Conn) {
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
