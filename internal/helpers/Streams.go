package helpers

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"io"
	"net"
	"os"
)

var red = color.New(color.FgRed)

func PrintBytes(ctx context.Context, reader io.ReadCloser) error {
	buff := make([]byte, 4096)

	defer reader.Close()

	for {
		select {
		case <-ctx.Done():
			return errors.New("context canceled")
		default:
			// Continue processing
			n, err := reader.Read(buff)

			if n > 0 {
				fmt.Print(string(buff[:n]))
			}

			if err != nil {
				if err == io.EOF {
					return nil
				}

				if errors.Is(err, net.ErrClosed) {
					return fmt.Errorf("connection closed")
				} else {
					return fmt.Errorf("error reading data: %w", err)
				}
			}
		}
	}
}

func PrintBytesDemux(ctx context.Context, reader io.ReadCloser) error {
	defer reader.Close()

	header := make([]byte, 8)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			_, err := io.ReadFull(reader, header)
			if err == io.EOF {
				return nil
			}

			if err != nil {
				return fmt.Errorf("logs streaming stopped", err)
			}

			stream := header[0]
			size := binary.BigEndian.Uint32(header[4:8])
			if size == 0 {
				continue
			}

			payload := make([]byte, size)
			_, err = io.ReadFull(reader, payload)
			if err != nil {
				return fmt.Errorf("reading payload: %w", err)
			}

			switch stream {
			case 1:
				os.Stdout.Write(payload)
			case 2:
				red.Fprintf(os.Stderr, "%s", payload)
			default:
				// unknown stream, ignore or log
			}
		}
	}
}
