package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts/iresponse"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"io"
	"net/http"
)

func (a *Api) Kind(c *gin.Context) {
	definition, err := io.ReadAll(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", err, nil))
	} else {
		dummy := v1.CommonDefinition{}
		err = dummy.FromJson(definition)

		if err != nil {
			c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", err, nil))
		} else {
			kindObj, ok := a.KindsRegistry[dummy.GetKind()]

			if !ok {
				c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", errors.New("invalid definition sent"), nil))
			} else {
				var response iresponse.Response

				switch c.Param("action") {
				case "apply":
					response, err = kindObj.Apply(authentication.NewUser(c.Request.TLS), definition, a.Config.NodeName)
					break
				case "state":
					response, err = kindObj.State(authentication.NewUser(c.Request.TLS), definition, a.Config.NodeName)
					break
				case "remove":
					response, err = kindObj.Delete(authentication.NewUser(c.Request.TLS), definition, a.Config.NodeName)
					break
				}

				c.JSON(response.HttpStatus, response)
			}
		}
	}
}
