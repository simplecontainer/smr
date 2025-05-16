package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/simplecontainer/smr/pkg/contracts/ievents"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/wss"
	"go.uber.org/zap"
	"sync"
	"time"
)

func (a *Api) Events(c *gin.Context) {
	w, r := c.Writer, c.Request
	lock := &sync.RWMutex{}

	conn, err := wssUpgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Log.Error("failed to upgrade WebSocket connection: ", zap.Error(err))
		return
	}
	defer conn.Close()

	a.Wss.Lock.Lock()
	ch := make(chan ievents.Event, 100)
	position := len(a.Wss.Channels)
	a.Wss.Channels[position] = ch
	a.Wss.Lock.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var closeOnce sync.Once

	go func(lock *sync.RWMutex) {
		defer func() {
			a.Wss.Lock.Lock()
			defer a.Wss.Lock.Unlock()

			closeOnce.Do(func() {
				close(ch)
			})

			delete(a.Wss.Channels, position)
		}()

		ListenEvents(ctx, a.Wss, position, conn, lock)
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
					logger.Log.Warn("failed to send ping: ", zap.Error(err))
					cancel()
					closeOnce.Do(func() { close(ch) })

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
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Log.Warn("websocket closed unexpectedly: ", zap.Error(err))
			} else {
				logger.Log.Info("websocket client disconnected normally")
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
			bytes, err := data.ToJSON()
			if err != nil {
				logger.Log.Error("failed to serialize event: ", zap.Error(err))
				continue
			}

			message, err := websocket.NewPreparedMessage(websocket.TextMessage, bytes)
			if err != nil {
				logger.Log.Error("failed to prepare WebSocket message: ", zap.Error(err))
				continue
			}

			lock.Lock()
			err = conn.WritePreparedMessage(message)
			lock.Unlock()

			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logger.Log.Warn("websocket write error: ", zap.Error(err))
				}
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
