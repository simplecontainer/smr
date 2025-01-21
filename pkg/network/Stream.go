package network

import (
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

func StreamHttp(reader io.ReadCloser, w gin.ResponseWriter) error {
	var bytes int
	var err error

	buff := make([]byte, 512)

	for {
		bytes, err = reader.Read(buff)

		if bytes == 0 || err == io.EOF {
			err = reader.Close()

			if err != nil {
				return err
			}

			w.(http.Flusher).Flush()
		} else {
			_, err = w.Write(buff[:bytes])

			if err != nil {
				return err
			}
		}
	}
}

func StreamByte(bytes []byte, w gin.ResponseWriter) error {
	var err error

	_, err = w.Write(bytes)

	if err != nil {
		return err
	}

	w.(http.Flusher).Flush()
	return nil
}

func StreamClose(w gin.ResponseWriter) {
	w.CloseNotify()
}
