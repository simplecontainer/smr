package tail

import (
	"context"
	"io"
	"sync"
)

type FileTailer struct {
	path       string
	follow     bool
	reader     *io.PipeReader
	writer     *io.PipeWriter
	done       chan struct{}
	mutex      sync.Mutex
	ctx        context.Context
	cancelFunc context.CancelFunc
	closed     bool
}
