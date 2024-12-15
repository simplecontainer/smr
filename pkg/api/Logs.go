package api

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hpcloud/tail"
	"github.com/simplecontainer/smr/pkg/contracts"
	"net/http"
	"os"
)

func (api *Api) Logs(c *gin.Context) {
	kind := c.Param("kind")
	group := c.Param("group")
	identifier := c.Param("identifier")

	w := c.Writer
	header := w.Header()
	header.Set("Transfer-Encoding", "chunked")
	header.Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := os.Stat("/tmp/%s.%s.%s.log"); errors.Is(err, os.ErrNotExist) {

	}

	t, err := tail.TailFile(fmt.Sprintf("/tmp/%s.%s.%s.log", kind, group, identifier),
		tail.Config{
			Follow: true,
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, contracts.Response{
			Explanation:      "",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})

		return
	}

	for line := range t.Lines {
		select {
		case <-w.CloseNotify():
			return
		default:
			w.Write([]byte(fmt.Sprintf("%s\n", line.Text)))
			w.(http.Flusher).Flush()
		}
	}
}
