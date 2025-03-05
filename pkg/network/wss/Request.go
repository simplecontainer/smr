package wss

import (
	"github.com/gorilla/websocket"
	"net/http"
	"strings"
)

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
