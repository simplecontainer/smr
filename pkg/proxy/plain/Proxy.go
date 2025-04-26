package plain

import (
	"context"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/stream"
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
	//pipeReader, pipeWriter := io.Pipe()
	//
	//go func() {
	//	defer pipeWriter.Close()
	//	buffer := make([]byte, 4096)
	//
	//	ticker := time.NewTicker(1 * time.Second)
	//	defer ticker.Stop()
	//
	//	for {
	//		select {
	//		case <-p.context.Done():
	//			fmt.Println("reader: context cancelled")
	//			return
	//		default:
	//			n, err := p.remote.Read(buffer)
	//			if n > 0 {
	//				if _, writeErr := pipeWriter.Write(buffer[:n]); writeErr != nil {
	//					fmt.Println("reader: write to pipe error:", writeErr)
	//					return
	//				}
	//			}
	//
	//			if err != nil {
	//				if err == io.EOF {
	//					fmt.Println("reader: remote closed cleanly")
	//				} else {
	//					fmt.Println("reader: remote read error:", err)
	//				}
	//				return
	//			}
	//		}
	//	}
	//}()

	go func() {
		<-p.context.Done()
		if p.remote != nil {
			_ = p.remote.Close()
			fmt.Println("proxy: closed remote connection")
		}
	}()

	err := stream.Stream(p.local, p.remote)
	if err != nil {
		fmt.Println("proxy: streaming error:", err)
		return err
	}

	return nil
}
