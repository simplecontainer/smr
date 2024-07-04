package api

import (
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/bootstrap"
	"net/http"
)

func (api *Api) Init(c *gin.Context) {
	bootstrap.DeleteProject("test", api.Config)
	bootstrap.CreateProject("test", api.Config)

	c.JSON(http.StatusOK, gin.H{
		"Starting project": "started",
	})
}
