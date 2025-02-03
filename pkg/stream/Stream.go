package stream

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hpcloud/tail"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/network"
	"io"
	"net/http"
)

func Stream(w gin.ResponseWriter, reader io.ReadCloser) error {
	return network.StreamHttp(reader, w)
}

func StreamRemote(w gin.ResponseWriter, URL string, client *client.Client) {
	resp, err := network.Raw(client.Http, URL, http.MethodGet, nil)

	if err != nil {
		Bye(w, err)
	} else {
		err = Stream(w, resp.Body)

		if err != nil {
			Bye(w, err)
		}
	}
}

func StreamTail(w gin.ResponseWriter, path string, follow bool) error {
	t, err := tail.TailFile(path,
		tail.Config{
			Follow: follow,
		},
	)

	if err != nil {
		return err
	}

	for line := range t.Lines {
		select {
		case <-w.CloseNotify():
			return nil
		default:
			err = network.StreamByte([]byte(fmt.Sprintf("%s\n", line.Text)), w)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func Bye(w gin.ResponseWriter, err error) {
	if err != io.EOF {
		err = network.StreamByte([]byte(err.Error()), w)
	}

	if err != nil {
		logger.Log.Error(err.Error())
	}

	network.StreamClose(w)
}
