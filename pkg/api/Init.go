package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (api *Api) Init(c *gin.Context) {
	api.Manager.DeleteProject("test")
	api.Manager.CreateProject("test")

	c.JSON(http.StatusOK, gin.H{
		"Starting project": "started",
	})
}
