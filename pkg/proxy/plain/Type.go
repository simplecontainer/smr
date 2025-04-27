package plain

import (
	"context"
	"io"
)

type Proxy struct {
	context context.Context
	cancel  context.CancelFunc
	client  chan error
	server  chan error
	local   io.Writer
	remote  io.ReadCloser
}
