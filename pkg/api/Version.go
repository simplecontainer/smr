package api

import (
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/network"
	"net/http"
)

func (api *Api) GetVersion(c *gin.Context) {
	c.JSON(http.StatusOK, common.Response(http.StatusOK, "", nil, network.ToJson(api.Version)))
}
