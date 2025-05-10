package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/wss"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func (c *Client) Events(ctx context.Context, cancel context.CancelFunc, handle func(context.Context, context.CancelFunc, func() error, *websocket.Conn) error) error {
	url := fmt.Sprintf("%s/events", c.Context.APIURL)

	conn, cancelWSS, err := wss.Request(ctx, c.Context.GetClient(), nil, url)
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}
	defer conn.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan

		err = cancelWSS()

		if err != nil {
			fmt.Println(err)
		}
	}()

	return handle(ctx, cancel, cancelWSS, conn)
}

func (c *Client) ReadEvents(ctx context.Context, conn *websocket.Conn, msgChannel chan<- []byte) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_, msg, err := conn.ReadMessage()

			if err == nil {
				msgChannel <- msg
				continue
			}

			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
				return err
			}

			var closeErr *websocket.CloseError
			if !errors.As(err, &closeErr) {
				return err // Not a close error, return original error
			}

			text := strings.TrimSpace(closeErr.Text)
			if text == io.EOF.Error() || closeErr.Code == websocket.CloseNormalClosure {
				return nil // Normal closure, no error
			}

			return err
		}
	}
}
