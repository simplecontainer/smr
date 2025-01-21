package network

import (
	"bytes"
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

		if err == io.EOF {
			err = reader.Close()

			if err != nil {
				return err
			}
		} else {
			_, err = w.Write(buff[:bytes])

			if err != nil {
				return err
			}

			w.(http.Flusher).Flush()
		}
	}
}

func StreamByte(b []byte, w gin.ResponseWriter) error {
	var read int
	var err error

	buff := make([]byte, 512)

	reader := bytes.NewReader(b)

	for {
		read, err = reader.Read(buff)

		if err == io.EOF {
			break
		}

		_, err = w.Write(buff[:read])

		if err != nil {
			return err
		}

		w.(http.Flusher).Flush()
	}

	return nil
}

func StreamClose(w gin.ResponseWriter) {
	w.CloseNotify()
}
