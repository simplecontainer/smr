package stream

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/hpcloud/tail"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/network"
)

func Stream(w gin.ResponseWriter, reader io.ReadCloser) error {
	if reader == nil {
		return errors.New("nil reader provided")
	}
	return network.StreamHttp(reader, w)
}

func StreamRemote(c *gin.Context, w gin.ResponseWriter, URL string, client *client.Client) error {
	if client == nil {
		return errors.New("nil client provided")
	}

	resp, err := network.Raw(client.Http, URL, http.MethodGet, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to remote endpoint: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("remote server error (status %d): %s", resp.StatusCode, string(body))
	}

	go func() {
		<-c.Request.Context().Done()
		resp.Body.Close()
	}()

	return Stream(w, resp.Body)
}

func StreamTail(c *gin.Context, w gin.ResponseWriter, path string, follow bool) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", path)
	}

	config := tail.Config{
		Follow:    follow,
		ReOpen:    follow, // Reopen the file if it's rotated and we're following
		MustExist: true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: io.SeekStart}, // Start from the top of the file
	}

	t, err := tail.TailFile(path, config)
	if err != nil {
		return fmt.Errorf("failed to tail file: %w", err)
	}
	defer t.Cleanup()

	ctx := c.Request.Context()
	wg := sync.WaitGroup{}
	streamErr := make(chan error, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				streamErr <- ctx.Err()
				return
			case line, ok := <-t.Lines:
				if !ok {
					// Channel is closed
					return
				}
				if line.Err != nil {
					streamErr <- line.Err
					return
				}
				if err := network.StreamByte([]byte(fmt.Sprintf("%s\n", line.Text)), w); err != nil {
					streamErr <- err
					return
				}
			}
		}
	}()

	go func() {
		wg.Wait()
		close(streamErr)
	}()

	err, ok := <-streamErr
	if ok {
		return err
	}
	return nil
}

func Bye(w gin.ResponseWriter, err error) {
	if err == nil || err == io.EOF {
		network.StreamClose(w)
		return
	}

	sendErr := network.StreamByte([]byte(fmt.Sprintf("Error: %s", err.Error())), w)
	if sendErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to send error to client: %v (original error: %v)\n", sendErr, err)
	}

	network.StreamClose(w)
}
