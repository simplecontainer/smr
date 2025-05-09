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

func (c *Client) Events(ctx context.Context, cancel context.CancelFunc, handle func(context.Context, context.CancelFunc, *websocket.Conn) error) error {
	url := fmt.Sprintf("%s/events", c.Context.APIURL)

	conn, err := wss.Request(c.Context.GetClient(), nil, url)
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}
	defer conn.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "context canceled")

		if err != nil {
			conn.Close()
		}

		err = conn.WriteMessage(websocket.CloseMessage, msg)

		if err != nil {
			conn.Close()
		}

		fmt.Println("context canceled")
		return
	}()

	return handle(ctx, cancel, conn)
}

func (c *Client) ReadEvents(ctx context.Context, conn *websocket.Conn, msgChannel chan<- []byte) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_, msg, err := conn.ReadMessage()

			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
					return err
				} else {
					var closeErr *websocket.CloseError

					if errors.As(err, &closeErr) {
						text := strings.TrimSpace(closeErr.Text)

						if text == io.EOF.Error() || closeErr.Code == websocket.CloseNormalClosure {
							return nil
						} else {
							return err
						}
					}
				}

				return err
			}

			msgChannel <- msg
		}
	}
}
