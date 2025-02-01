package api

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"io"
	"net/http"
)

func (api *Api) Apply(c *gin.Context) {
	jsonData, err := io.ReadAll(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", err, nil))
	} else {
		data := make(map[string]interface{})
		err = json.Unmarshal(jsonData, &data)

		if err != nil {
			c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", err, nil))
		} else {
			kind := data["kind"].(string)
			kindObj, ok := api.KindsRegistry[kind]

			if !ok {
				c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", errors.New("invalid definition sent"), nil))
			} else {
				var response contracts.Response

				response, err = kindObj.Apply(authentication.NewUser(c.Request.TLS), jsonData, api.Config.NodeName)
				c.JSON(response.HttpStatus, response)
			}
		}
	}
}

func (api *Api) Delete(c *gin.Context) {
	jsonData, err := io.ReadAll(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", err, nil))
	} else {
		data := make(map[string]interface{})
		err = json.Unmarshal(jsonData, &data)

		if err != nil {
			c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", err, nil))
		} else {
			kind := data["kind"].(string)
			kindObj, ok := api.KindsRegistry[kind]

			if !ok {
				c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", errors.New("invalid definition sent"), nil))
			} else {
				var response contracts.Response

				response, err = kindObj.Delete(authentication.NewUser(c.Request.TLS), jsonData, api.Config.NodeName)
				c.JSON(response.HttpStatus, response)
			}
		}
	}
}

func (api *Api) Compare(c *gin.Context) {
	jsonData, err := io.ReadAll(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", err, nil))
	} else {
		data := make(map[string]interface{})
		err = json.Unmarshal(jsonData, &data)

		if err != nil {
			c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", err, nil))
		} else {
			kind := data["kind"].(string)
			kindObj, ok := api.KindsRegistry[kind]

			if !ok {
				c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", errors.New("invalid definition sent"), nil))
			} else {
				var response contracts.Response

				response, err = kindObj.Compare(authentication.NewUser(c.Request.TLS), jsonData)
				c.JSON(response.HttpStatus, response)
			}
		}
	}
}
