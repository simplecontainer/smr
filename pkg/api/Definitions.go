package api

import (
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/helpers"
	"net/http"
)

func (api *Api) Definition(c *gin.Context) {
	definition, _ := helpers.Definitions(c.Param("definition"))

	c.JSON(http.StatusOK, contracts.ResponseImplementation{
		HttpStatus:       http.StatusOK,
		Explanation:      "definition in yaml format",
		ErrorExplanation: "",
		Error:            true,
		Success:          false,
		Data:             definition,
	})
}

func (api *Api) Definitions(c *gin.Context) {
	c.JSON(http.StatusOK, contracts.ResponseImplementation{
		HttpStatus:       http.StatusOK,
		Explanation:      "definitions loaded on the server",
		ErrorExplanation: "",
		Error:            true,
		Success:          false,
		Data:             api.Manager.Kinds,
	})
}
