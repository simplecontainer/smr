package stream

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

func Byte(b []byte, w gin.ResponseWriter) error {
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
