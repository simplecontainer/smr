package exec

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/logger"
	"go.uber.org/zap"
	"io"
	"net"
	"sync"
	"time"
)

func Create(c context.Context, cancel context.CancelFunc, clientConn *websocket.Conn, container platforms.IContainer,
	command []string, interactive bool, height string, width string) (*Session, error) {
	execID, reader, execConn, err := container.Exec(c, command, interactive, height, width)

	if err != nil {
		if clientConn != nil {
			// If used with websocket - inform the client on the other end of socket
			clientConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, err.Error()))
		}
		return nil, err
	}

	return &Session{
		ID:         execID,
		Reader:     reader,
		Conn:       &execConn,
		ClientConn: clientConn,
		Container:  container,
		context:    c,
		cancel:     cancel,
		engine:     make(chan error),
		client:     make(chan error),
	}, nil
}

func (s *Session) Exec() error {
	logger.Log.Debug("exec session started", zap.String("execID", s.ID))

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		if err := input(s.context, s.ClientConn, s.Conn, s.Reader, s.ID, s.Container); err != nil {
			logger.Log.Info("client to remote exec ended", zap.Error(err))

			select {
			case s.client <- err:
			case <-s.context.Done():
			}
		}

		return
	}()

	go func() {
		defer wg.Done()

		if err := output(s.context, s.ClientConn, s.Conn, s.Reader); err != nil {
			logger.Log.Info("server exec to remote client ended", zap.Error(err))

			select {
			case s.engine <- err:
			case <-s.context.Done():
			}
		}

		return
	}()

	timeout := time.After(60 * time.Minute)

	var reason error

	for {
		select {
		case <-timeout:
			s.cancel()
		case err := <-s.engine:
			reason = err
			(*s.ClientConn).WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, err.Error()))
			s.cancel()
		case err := <-s.client:
			reason = err
			(*s.Conn).Close()
			s.cancel()
		case <-s.context.Done():
			wg.Wait()
			return reason
		}
	}
}

func input(ctx context.Context, websocketConn *websocket.Conn, execConn *net.Conn, reader *bufio.Reader, sessionID string, container platforms.IContainer) error {
	for {
		select {
		case <-ctx.Done():
			return websocketConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "context done"))
		default:
			messageType, message, err := websocketConn.ReadMessage()

			if err != nil {
				var netErr net.Error

				if errors.As(err, &netErr) && netErr.Timeout() {
					continue // Just try again after timeout
				}

				if closeErr := handleCloseMaybe(execConn, err); closeErr != nil {
					return closeErr
				}

				if err == io.EOF {
					defer func(conn net.Conn) {
						err := conn.Close()
						if err != nil {
							logger.Log.Error("failed to close exec connection", zap.Error(err))
						}
					}(*execConn)
				}

				return err
			}

			switch messageType {
			case websocket.BinaryMessage:
				fmt.Println("binary", string(message))
				_, err = (*execConn).Write(message)

				if err != nil {
					return err
				}
				break
			case websocket.TextMessage:
				fmt.Println("text", string(message))
				ctrl, err := UnmarshalControl(message)
				if err != nil {
					return err
				}

				err = control(container, sessionID, ctrl)
				if err != nil {
					return err
				}
				break
			default:
				break
			}
		}
	}
}

func output(ctx context.Context, websocketConn *websocket.Conn, execConn *net.Conn, reader *bufio.Reader) error {
	buf := make([]byte, 4096)
	frameBuf := make([]byte, 0)

	for {
		select {
		case <-ctx.Done():
			return websocketConn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "context done"),
			)
		default:
			n, err := reader.Read(buf)
			if n > 0 {
				frameBuf = append(frameBuf, buf[:n]...)

				for {
					if len(frameBuf) < 8 {
						break
					}

					frameSize := binary.BigEndian.Uint32(frameBuf[4:8])
					totalFrameLen := 8 + int(frameSize)

					if len(frameBuf) < totalFrameLen {
						break
					}

					frame := frameBuf[:totalFrameLen]
					if err := websocketConn.WriteMessage(websocket.BinaryMessage, frame); err != nil {
						return err
					}

					frameBuf = frameBuf[totalFrameLen:]
				}
			}

			if err != nil {
				if err == io.EOF {
					return websocketConn.WriteMessage(
						websocket.CloseMessage,
						websocket.FormatCloseMessage(websocket.CloseNormalClosure, io.EOF.Error()),
					)
				}
				return err
			}
		}
	}
}

func control(container platforms.IContainer, sessionID string, ctrl *Control) error {
	switch ctrl.Type {
	case RESIZE_TYPE:
		resize, err := ctrl.DecodeResize()
		if err != nil {
			return err
		}

		return container.ExecResize(sessionID, resize.Width, resize.Height)
	}

	return errors.New("unsupported control message")
}

func handleCloseMaybe(target *net.Conn, err error) error {
	var parsed *websocket.CloseError

	if errors.As(err, &parsed) {
		defer func(conn net.Conn) {
			err := conn.Close()
			if err != nil {
				logger.Log.Error("failed to close exec connection", zap.Error(err))
			}
		}(*target)

		if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
			return errors.New("abnormal closure")
		} else {
			return errors.New("normal closure")
		}
	} else {
		return nil
	}
}
