package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/simplecontainer/smr/pkg/contracts/ievents"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/wss"
	"net/http"
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
		return
	}

	api.Wss.Lock.Lock()
	api.Wss.Channels = append(api.Wss.Channels, make(chan ievents.Event))
	position := len(api.Wss.Channels) - 1
	api.Wss.Lock.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	go ListenEvents(ctx, api.Wss, position, conn)

	defer conn.Close()
	for {
		_, _, err = conn.ReadMessage()

		if err != nil {
			cancel()
			break
		}
	}
}

func ListenEvents(ctx context.Context, wss *wss.WebSockets, position int, conn *websocket.Conn) {
	for {
		select {
		case data := <-wss.Channels[position]:
			bytes, err := data.ToJson()

			if err != nil {
				logger.Log.Error(err.Error())
			} else {
				var message *websocket.PreparedMessage

				message, err = websocket.NewPreparedMessage(websocket.TextMessage, bytes)

				if err != nil {
					logger.Log.Error(err.Error())
				}

				err = conn.WritePreparedMessage(message)

				if err != nil {
					logger.Log.Error(err.Error())
				}
			}
		case <-ctx.Done():
			wss.Lock.Lock()
			if position >= 0 && position < len(wss.Channels) {
				wss.Channels = append(wss.Channels[:position], wss.Channels[position+1:]...)
			}
			wss.Lock.Unlock()
			return
		}
	}
}
