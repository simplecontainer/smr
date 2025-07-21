package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/contracts/iformat"
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/wss"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func (c *Client) Events(ctx context.Context, cancel context.CancelFunc, event string, identifier string, handle func(context.Context, context.CancelFunc, string, string, func() error, *websocket.Conn) error) error {
	url := fmt.Sprintf("%s/events", c.Context.APIURL)

	conn, cancelWSS, err := wss.Request(ctx, c.Context.GetHTTPClient(), nil, url)
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
			helpers.PrintAndExit(err, 1)
		}
	}()

	return handle(ctx, cancel, event, identifier, cancelWSS, conn)
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

func (c *Client) Tracker(ctx context.Context, cancel context.CancelFunc, waitForEvent string, identifier string, cancelWSS func() error, conn *websocket.Conn) error {
	defer cancel()
	msgChannel := make(chan []byte)

	go func() {
		err := c.ReadEvents(ctx, conn, msgChannel)

		if err != nil {
			helpers.PrintAndExit(err, 1)
		}

		close(msgChannel)
		return
	}()

	var format iformat.Format
	var err error

	if identifier != "" {
		format, err = f.Build(identifier, c.Group)
		if err != nil {
			helpers.PrintAndExit(err, 1)
		}
	}

	if waitForEvent != "" {
		for msg := range msgChannel {
			var event events.Event

			err := json.Unmarshal(msg, &event)

			if err != nil {
				fmt.Println(err)
				continue
			}

			if event.Type == waitForEvent {
				if identifier != "" {
					if event.IsOfFormat(format) {

						err = cancelWSS()

						if err != nil {
							fmt.Println(err)
						}
					}
				} else {
					err = cancelWSS()

					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}
	} else {
		for msg := range msgChannel {
			fmt.Printf("event: %s\n", msg)
		}
	}

	return nil
}
