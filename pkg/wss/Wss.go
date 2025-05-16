package wss

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/simplecontainer/smr/pkg/contracts/ievents"
	"net/http"
	"strings"
	"sync"
	"time"
)

func New() *WebSockets {
	return &WebSockets{
		Channels:    make(map[int]chan ievents.Event, 0),
		Lock:        &sync.RWMutex{},
		Connections: make(map[string]*ConnectionControl),
	}
}

func Request(ctx context.Context, client *http.Client, headers http.Header, url string) (*websocket.Conn, func() error, error) {
	wsURL := strings.Replace(url, "https://", "wss://", 1)

	dialer := websocket.DefaultDialer

	dialer.Jar = nil
	dialer.Proxy = nil
	dialer.TLSClientConfig = client.Transport.(*http.Transport).TLSClientConfig

	connCtx, cancel := context.WithCancel(ctx)

	conn, _, err := dialer.DialContext(connCtx, wsURL, headers)
	if err != nil {
		cancel() // Clean up if connection fails
		return nil, nil, err
	}

	connMutex := &sync.Mutex{}
	isClosed := false

	closeFunc := func() error {
		connMutex.Lock()
		defer connMutex.Unlock()

		if isClosed {
			return nil
		}

		isClosed = true

		_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

		closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "connection closed")
		err = conn.WriteMessage(websocket.CloseMessage, closeMsg)

		cancel()

		time.Sleep(100 * time.Millisecond)
		return nil
	}

	go func() {
		<-connCtx.Done()
		connMutex.Lock()
		if !isClosed {
			connMutex.Unlock()
			_ = closeFunc()
			return
		}
		connMutex.Unlock()
	}()

	return conn, closeFunc, nil
}
