package wss

import (
	"github.com/gorilla/websocket"
	"github.com/simplecontainer/smr/pkg/logger"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
)

func StreamRemote(client *http.Client, clientConn *websocket.Conn, url string) (*websocket.Conn, error) {
	wsURL := strings.Replace(url, "https://", "wss://", 1)

	dialer := websocket.DefaultDialer

	dialer.Jar = nil
	dialer.Proxy = nil
	dialer.TLSClientConfig = client.Transport.(*http.Transport).TLSClientConfig

	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		return nil, err
	}

	go proxyMessages(clientConn, conn)
	proxyMessages(conn, clientConn)

	return nil, nil
}

func proxyMessages(src, dest *websocket.Conn) {
	for {
		messageType, msg, err := src.ReadMessage()
		if err != nil {
			if err != io.EOF {
				logger.Log.Error("wss read proxy error:", zap.Error(err))
			}
			break
		}

		err = dest.WriteMessage(messageType, msg)
		if err != nil {
			logger.Log.Error("wss write proxy error:", zap.Error(err))
			break
		}
	}
}
