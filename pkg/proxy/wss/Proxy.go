package wss

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/simplecontainer/smr/pkg/logger"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

func New(ctx context.Context, cancel context.CancelFunc, client *http.Client, requestHeader http.Header, clientConn *websocket.Conn, url string) (*Proxy, error) {
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

	headers := http.Header{}

	for k, v := range requestHeader {
		if k == "User-Agent" || k == "Connection" || k == "Upgrade" || strings.Contains(strings.ToLower(k), "sec-websocket") {
			continue
		}

		headers[k] = v
	}

	headers.Add("User-Agent", "smr-wss-proxy")

	logger.Log.Info("dialing remote WebSocket", zap.String("url", wsURL))

	serverConn, _, err := dialer.Dial(wsURL, headers)

	if err != nil {
		logger.Log.Info("failed to dial websocket", zap.Error(err))
		clientConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, err.Error()))
		return nil, err
	}

	return &Proxy{
		context:    ctx,
		cancel:     cancel,
		clientConn: clientConn,
		serverConn: serverConn,
		client:     make(chan error),
		server:     make(chan error),
	}, nil
}

func (p *Proxy) WSS() error {
	defer p.cancel()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if err := forward(p.context, p.clientConn, p.serverConn); err != nil {
			logger.Log.Info("client to remote proxy ended", zap.Error(err))

			select {
			case p.client <- err:
			case <-p.context.Done():
			}
		}
	}()

	go func() {
		defer wg.Done()
		if err := forward(p.context, p.serverConn, p.clientConn); err != nil {
			logger.Log.Info("remote to client proxy ended", zap.Error(err))

			select {
			case p.server <- err:
			case <-p.context.Done():
			}
		}
	}()

	var reason error

	for {
		select {
		case err := <-p.client:
			reason = err
			(*p.serverConn).WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, err.Error()))
			p.cancel()
		case err := <-p.server:
			reason = err
			(*p.clientConn).WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, err.Error()))
			p.cancel()
		case <-p.context.Done():
			wg.Wait()
			return reason
		}
	}
}

func forward(ctx context.Context, source *websocket.Conn, target *websocket.Conn) error {
	for {
		select {
		case <-ctx.Done():
			return target.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "context done"))
		default:
			messageType, message, err := source.ReadMessage()

			if err != nil {
				if closeErr := handleCloseMaybe(target, err); closeErr != nil {
					return closeErr
				}

				if err == io.EOF {
					return target.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "eof"))
				}

				return err
			}

			switch messageType {
			case websocket.PingMessage, websocket.PongMessage:
				if err := target.WriteMessage(messageType, nil); err != nil {
					logger.Log.Error("error sending pong", zap.Error(err))
					return err
				}
				break
			default:
				if err := target.WriteMessage(messageType, message); err != nil {
					logger.Log.Error("error writing to websocket", zap.Error(err))
					return err
				}
				break
			}
		}
	}
}

func handleCloseMaybe(target *websocket.Conn, err error) error {
	var parsed *websocket.CloseError

	if errors.As(err, &parsed) {
		defer target.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(parsed.Code, parsed.Text))
		return errors.New(parsed.Text)
	} else {
		return nil
	}
}
