package wss

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/simplecontainer/smr/pkg/logger"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

func StreamRemote(client *http.Client, clientConn *websocket.Conn, url string) error {
	wsURL := url

	if strings.HasPrefix(wsURL, "http://") {
		wsURL = "ws://" + strings.TrimPrefix(wsURL, "http://")
	} else if strings.HasPrefix(wsURL, "https://") {
		wsURL = "wss://" + strings.TrimPrefix(wsURL, "https://")
	}

	dialer := &websocket.Dialer{
		TLSClientConfig:  client.Transport.(*http.Transport).TLSClientConfig,
		HandshakeTimeout: 10 * time.Second,
	}

	requestHeader := http.Header{}
	requestHeader.Add("User-Agent", "SMR-WSS-Client")

	logger.Log.Debug("Dialing remote WebSocket", zap.String("url", wsURL))
	conn, resp, err := dialer.Dial(wsURL, requestHeader)
	if err != nil {
		if resp != nil {
			logger.Log.Error("WebSocket dial failed",
				zap.Error(err),
				zap.Int("statusCode", resp.StatusCode),
				zap.String("status", resp.Status))
		} else {
			logger.Log.Error("WebSocket dial failed with no response", zap.Error(err))
		}
		return err
	}
	defer conn.Close()

	logger.Log.Debug("Connected to remote WebSocket")

	var wg sync.WaitGroup
	wg.Add(2)

	errChan := make(chan error, 2)

	go func() {
		defer wg.Done()
		if err := proxyWebSocketMessages(clientConn, conn); err != nil {
			logger.Log.Debug("Client to remote proxy ended", zap.Error(err))
			select {
			case errChan <- err:
			default:
			}
		}
	}()

	go func() {
		defer wg.Done()
		if err := proxyWebSocketMessages(conn, clientConn); err != nil {
			logger.Log.Debug("Remote to client proxy ended", zap.Error(err))
			select {
			case errChan <- err:
			default:
			}
		}
	}()

	doneChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneChan)
	}()

	select {
	case err := <-errChan:
		if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
			logger.Log.Debug("WebSocket closed normally")
			return nil
		}
		logger.Log.Error("WebSocket proxy error", zap.Error(err))
		return err
	case <-doneChan:
		logger.Log.Debug("WebSocket proxy completed normally")
		return nil
	}
}

func proxyWebSocketMessages(src, dest *websocket.Conn) error {
	for {
		messageType, message, err := src.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				logger.Log.Debug("WebSocket closed normally")

				var closeErr *websocket.CloseError

				if errors.As(err, &closeErr) {
					fmt.Printf(strings.TrimSpace(closeErr.Text))
				} else {
					fmt.Printf("invalid response from the server")
				}

				dest.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, closeErr.Text))
				return nil
			}

			if err == io.EOF {
				logger.Log.Debug("WebSocket EOF")
				return nil
			}

			logger.Log.Error("Error reading from WebSocket", zap.Error(err))
			return err
		}

		if messageType == websocket.PingMessage {
			if err := src.WriteMessage(websocket.PongMessage, nil); err != nil {
				logger.Log.Error("Error sending pong", zap.Error(err))
			}
			continue
		}

		if err := dest.WriteMessage(messageType, message); err != nil {
			logger.Log.Error("Error writing to WebSocket", zap.Error(err))
			return err
		}
	}
}
