package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/contracts"
	"io"
	"net/http"
	"time"
)

func (api *Api) EtcdPut(c *gin.Context) {
	timeout, err := time.ParseDuration("20s")

	if err != nil {
		c.JSON(http.StatusBadRequest, contracts.Response{
			Explanation:      "",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})

		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	var body []byte
	body, err = io.ReadAll(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusBadRequest, contracts.Response{
			Explanation:      "",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})

		cancel()
		return
	}

	_, err = api.Cluster.EtcdClient.Put(ctx, c.Param("key"), string(body))
	cancel()

	c.JSON(http.StatusOK, contracts.Response{
		Explanation:      "",
		ErrorExplanation: "all goodies",
		Error:            false,
		Success:          true,
		Data:             nil,
	})
}
func (api *Api) EtcdDelete(c *gin.Context) {}
