package exec

import (
	"bufio"
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"net"
)

type Session struct {
	ID         string
	Container  platforms.IContainer
	Reader     *bufio.Reader
	Conn       *net.Conn
	ClientConn *websocket.Conn
	context    context.Context
	cancel     context.CancelFunc
	engine     chan error
	client     chan error
}

type Control struct {
	Type int
	Data json.RawMessage
}

type Resize struct {
	Width  int
	Height int
}

const RESIZE_TYPE = 1
