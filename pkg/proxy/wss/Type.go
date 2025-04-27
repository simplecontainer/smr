package wss

import (
	"context"
	"github.com/gorilla/websocket"
)

type Proxy struct {
	context    context.Context
	cancel     context.CancelFunc
	clientConn *websocket.Conn
	serverConn *websocket.Conn
	client     chan error
	server     chan error
}
