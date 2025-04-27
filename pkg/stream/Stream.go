package stream

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"os"
)

func Stream(w io.Writer, reader io.ReadCloser) error {
	if reader == nil {
		return errors.New("nil reader provided")
	}

	return Http(reader, w)
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
