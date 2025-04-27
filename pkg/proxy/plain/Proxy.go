package plain

import (
	"context"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/stream"
	"go.uber.org/zap"
	"io"
	"net/http"
)

func Create(ctx context.Context, cancel context.CancelFunc, local io.Writer, remote io.ReadCloser) *Proxy {
	return &Proxy{
		context: ctx,
		cancel:  cancel,
		client:  make(chan error),
		server:  make(chan error),
		local:   local,
		remote:  remote,
	}
}

func Dial(ctx context.Context, cancel context.CancelFunc, client *http.Client, URL string) (io.ReadCloser, error) {
	if client == nil {
		return nil, errors.New("nil client provided")
	}

	resp, err := network.Raw(client, URL, http.MethodGet, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to remote endpoint: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("remote server error (status %d): %s", resp.StatusCode, string(body))
	}

	return resp.Body, nil
}

func (p *Proxy) Proxy() error {
	go func() {
		<-p.context.Done()
		if p.remote != nil {
			_ = p.remote.Close()
			logger.Log.Info("proxy: closed remote connection")
		}
	}()

	err := stream.Stream(p.local, p.remote)
	if err != nil {
		logger.Log.Info("proxy: streaming error:", zap.Error(err))
		return err
	}

	return nil
}
