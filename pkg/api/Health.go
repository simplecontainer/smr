package api

import (
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/contracts/iresponse"
	"net/http"
)

func (a *Api) Health(c *gin.Context) {
	c.JSON(http.StatusOK, &iresponse.Response{
		HttpStatus:       http.StatusOK,
		Explanation:      "agent is healthy",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             nil,
	})
}
