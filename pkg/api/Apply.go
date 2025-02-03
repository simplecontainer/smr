package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"io"
	"net/http"
)

func (api *Api) Kind(c *gin.Context) {
	definition, err := io.ReadAll(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", err, nil))
	} else {
		dummy := v1.CommonDefinition{}
		err = dummy.FromJson(definition)

		if err != nil {
			c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", err, nil))
		} else {
			kindObj, ok := api.KindsRegistry[dummy.GetKind()]

			if !ok {
				c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", errors.New("invalid definition sent"), nil))
			} else {
				var response contracts.Response

				switch c.Param("action") {
				case "apply":
					response, err = kindObj.Apply(authentication.NewUser(c.Request.TLS), definition, api.Config.NodeName)
					break
				case "delete":
					response, err = kindObj.Delete(authentication.NewUser(c.Request.TLS), definition, api.Config.NodeName)
					break
				case "compare":
					response, err = kindObj.Compare(authentication.NewUser(c.Request.TLS), definition)
					break
				}

				c.JSON(response.HttpStatus, response)
			}
		}
	}
}
