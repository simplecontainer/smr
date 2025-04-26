package stream

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hpcloud/tail"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/network"
	"io"
	"net/http"
	"os"
)

func Stream(w io.Writer, reader io.ReadCloser) error {
	if reader == nil {
		return errors.New("nil reader provided")
	}

	return Http(reader, w)
}

func Remote(c *gin.Context, URL string, client *client.Client) error {
	if client == nil {
		return errors.New("nil client provided")
	}

	resp, err := network.Raw(client.Http, URL, http.MethodGet, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to remote endpoint: %w", err)
	}

	//defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("remote server error (status %d): %s", resp.StatusCode, string(body))
	}

	return Stream(c.Writer, resp.Body)
}

func Tail(c *gin.Context, w gin.ResponseWriter, path string, follow bool) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", path)
	}

	config := tail.Config{
		Follow:    follow,
		ReOpen:    follow,
		MustExist: true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: io.SeekStart},
	}

	t, err := tail.TailFile(path, config)
	if err != nil {
		return fmt.Errorf("failed to tail file: %w", err)
	}
	defer t.Cleanup()

	ctx := c.Request.Context()
	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)
		for {
			select {
			case <-ctx.Done():
				t.Stop()
				errCh <- ctx.Err()

				logger.Log.Info("context closed - tail dies")

				return
			case line, ok := <-t.Lines:
				if !ok {
					errCh <- nil
					return
				}

				if line.Err != nil {
					errCh <- line.Err
					return
				}

				if err := Byte([]byte(line.Text+"\n"), w); err != nil {
					errCh <- err
					return
				}
			}
		}
	}()

	return <-errCh
}

func Bye(w gin.ResponseWriter, err error) {
	if err == nil || err == io.EOF {
		Close(w)
		return
	}

	sendErr := Byte([]byte(fmt.Sprintf("Error: %s", err.Error())), w)
	if sendErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to send error to client: %v (original error: %v)\n", sendErr, err)
	}

	Close(w)
}

func ByeWithStatus(w gin.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	Bye(w, err)
}

func Close(w gin.ResponseWriter) {
	w.CloseNotify()
}
