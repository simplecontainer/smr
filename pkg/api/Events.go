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
	Subprotocols: []string{
		"Authorization",
	},
}

func (api *Api) Events(c *gin.Context) {
	w, r := c.Writer, c.Request
	lock := &sync.RWMutex{}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Log.Error("Failed to upgrade WebSocket connection: ", zap.Error(err))
		return
	}
	defer conn.Close()

	api.Wss.Lock.Lock()
	ch := make(chan ievents.Event, 100)
	position := len(api.Wss.Channels)
	api.Wss.Channels[position] = ch
	api.Wss.Lock.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var closeOnce sync.Once

	go func(lock *sync.RWMutex) {
		defer func() {
			api.Wss.Lock.Lock()
			defer api.Wss.Lock.Unlock()

			closeOnce.Do(func() {
				close(ch)
			})

			delete(api.Wss.Channels, position)
		}()

		ListenEvents(ctx, api.Wss, position, conn, lock)
	}(lock)

	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	go func() {
		for {
			select {
			case <-pingTicker.C:
				lock.Lock()
				err := conn.WriteMessage(websocket.PingMessage, []byte{})
				lock.Unlock()

				if err != nil {
					logger.Log.Warn("Failed to send ping: ", zap.Error(err))
					cancel()
					closeOnce.Do(func() { close(ch) })
					lock.Unlock()

					return
				}
			case <-ctx.Done():
				return
			}
		}
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

func ListenEvents(ctx context.Context, wss *wss.WebSockets, position int, conn *websocket.Conn, lock *sync.RWMutex) {
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

			lock.Lock()
			err = conn.WritePreparedMessage(message)
			lock.Unlock()

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
