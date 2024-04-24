package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (api *Api) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"healthy": "true",
		"project": api.Runtime.PROJECT,
	})
}
