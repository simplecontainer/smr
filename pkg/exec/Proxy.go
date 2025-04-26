package exec

import (
	"bufio"
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/logger"
	"go.uber.org/zap"
	"io"
	"net"
	"sync"
	"time"
)

type Session struct {
	ID         string
	Reader     *bufio.Reader
	Conn       *net.Conn
	ClientConn *websocket.Conn
	context    context.Context
	cancel     context.CancelFunc
	engine     chan error
	client     chan error
}

func Create(c context.Context, cancel context.CancelFunc, clientConn *websocket.Conn, container platforms.IContainer, command []string, interactive bool) (*Session, error) {
	execID, reader, execConn, err := container.Exec(c, command, interactive)

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
		context:    c,
		cancel:     cancel,
		engine:     make(chan error),
		client:     make(chan error),
	}, nil
}

func (s *Session) Exec() error {
	logger.Log.Debug("interactive session started", zap.String("execID", s.ID))

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		if err := input(s.context, s.ClientConn, s.Conn, s.Reader); err != nil {
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

func input(ctx context.Context, websocketConn *websocket.Conn, execConn *net.Conn, reader *bufio.Reader) error {
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
			case websocket.TextMessage, websocket.BinaryMessage:
				_, err = (*execConn).Write(message)

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
	for {
		select {
		case <-ctx.Done():
			return websocketConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "context done"))
		default:
			buf := make([]byte, 4096)

			n, err := reader.Read(buf)
			if err != nil {
				if err != io.EOF {
					return websocketConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "eof"))
				}

				return err
			}

			if n > 0 {
				err = websocketConn.WriteMessage(websocket.BinaryMessage, buf[:n])

				if err != nil {
					return err
				}
			}
		}
	}
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
