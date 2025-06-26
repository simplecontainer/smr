package exec

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gorilla/websocket"
	"io"
	"os"
	"strings"
)

func Write(ctx context.Context, conn *websocket.Conn) error {
	reader := bufio.NewReaderSize(os.Stdin, 1)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			input, err := reader.ReadByte()

			if err != nil {
				return err
			}

			err = conn.WriteMessage(websocket.BinaryMessage, []byte{input})

			if err != nil {
				return err
			}
		}
	}
}
func Read(ctx context.Context, conn *websocket.Conn) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			_, msg, err := conn.ReadMessage()

			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
					return err
				} else {
					var closeErr *websocket.CloseError

					if errors.As(err, &closeErr) {
						text := strings.TrimSpace(closeErr.Text)

						if text == io.EOF.Error() {
							return nil
						} else {
							os.Stdout.Write([]byte(text))
						}
					}

					return nil
				}
			}

			var outBuf, errBuf bytes.Buffer
			_, err = stdcopy.StdCopy(&outBuf, &errBuf, bytes.NewBuffer(msg))

			if err != nil {
				return err
			}

			if outBytes := outBuf.Bytes(); len(outBytes) > 0 {
				if _, err := os.Stdout.Write(outBytes); err != nil {
					return err
				}
			}

			if errBytes := errBuf.Bytes(); len(errBytes) > 0 {
				if _, err := os.Stderr.Write(errBytes); err != nil {
					return err
				}
			}
		}
	}
}
