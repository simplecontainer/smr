package stream

import (
	"io"
	"net/http"
	"strings"
)

func Http(reader io.ReadCloser, w io.Writer) error {
	buff := make([]byte, 4096)

	for {
		bytes, err := reader.Read(buff)

		if bytes > 0 {
			if _, writeErr := w.Write(buff[:bytes]); writeErr != nil {
				reader.Close()
				return writeErr
			}
		}

		w.(http.Flusher).Flush()

		if err != nil {
			if err == io.EOF {
				reader.Close()
				return nil
			}

			if err == io.ErrClosedPipe || strings.Contains(err.Error(), "closed") {
				return nil
			}

			reader.Close()
			return err
		}

		// Zero-byte read without error - this is unusual but valid
		// Continue the loop to try reading again
	}
}
