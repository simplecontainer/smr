package exec

import (
	"bufio"
	"context"
	"github.com/gorilla/websocket"
	"net"
)

type Session struct {
	ID         string
	Reader     *bufio.Reader
	Conn       *net.Conn
	ClientConn *websocket.Conn
	context    context.Context
	cancel     context.CancelFunc
	engine     chan error
	client     chan error
}
