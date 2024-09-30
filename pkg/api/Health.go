package api

import (
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/httpcontract"
	"net/http"
)

func (api *Api) Health(c *gin.Context) {
	c.JSON(http.StatusOK, &httpcontract.ResponseImplementation{
		HttpStatus:       http.StatusOK,
		Explanation:      "agent is healthy",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             nil,
	})
}
