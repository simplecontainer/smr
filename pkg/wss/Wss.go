package wss

import (
	"github.com/gorilla/websocket"
	"github.com/simplecontainer/smr/pkg/contracts/ievents"
	"net/http"
	"strings"
	"sync"
)

func New() *WebSockets {
	return &WebSockets{
		Channels: make(map[int]chan ievents.Event, 0),
		Lock:     &sync.RWMutex{},
	}
}

func Request(client *http.Client, url string) (*websocket.Conn, error) {
	wsURL := strings.Replace(url, "https://", "wss://", 1)

	dialer := websocket.DefaultDialer

	dialer.Jar = nil
	dialer.Proxy = nil
	dialer.TLSClientConfig = client.Transport.(*http.Transport).TLSClientConfig

	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
