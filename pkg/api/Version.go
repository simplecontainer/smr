package api

import (
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/contracts"
	"net/http"
)

func (api *Api) Version(c *gin.Context) {
	c.JSON(http.StatusOK, contracts.ResponseOperator{
		Explanation:      "server version",
		ErrorExplanation: "",
		Error:            false,
		Success:          false,
		Data:             map[string]any{"ServerVersion": api.VersionServer},
	})
}
