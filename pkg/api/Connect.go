package api

import (
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/contracts"
	"net/http"
)

func (api *Api) Connect(c *gin.Context) {
	c.JSON(http.StatusOK, &contracts.ResponseImplementation{
		HttpStatus:       http.StatusOK,
		Explanation:      "connection was success",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data: map[string][]string{
			"domains": api.Config.Domains,
			"ips":     api.Config.IPs,
		},
	})
}
