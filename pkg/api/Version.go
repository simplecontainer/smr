package api

import (
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/contracts/iresponse"
	"github.com/simplecontainer/smr/pkg/network"
	"net/http"
)

func (api *Api) Version(c *gin.Context) {
	c.JSON(http.StatusOK, iresponse.Response{
		Explanation:      "server version",
		ErrorExplanation: "",
		Error:            false,
		Success:          false,
		Data:             network.ToJson(map[string]any{"ServerVersion": api.VersionServer}),
	})
}
