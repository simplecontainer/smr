package api

import (
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/contracts"
	"net/http"
)

func (api *Api) Health(c *gin.Context) {
	c.JSON(http.StatusOK, &contracts.Response{
		HttpStatus:       http.StatusOK,
		Explanation:      "agent is healthy",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             nil,
	})
}
