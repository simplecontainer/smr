package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts"
	"io"
	"net/http"
)

func (api *Api) Apply(c *gin.Context) {
	jsonData, err := io.ReadAll(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusBadRequest, contracts.ResponseImplementation{
			HttpStatus:       http.StatusBadRequest,
			Explanation:      "invalid definition sent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		})
	} else {
		data := make(map[string]interface{})

		err = json.Unmarshal(jsonData, &data)
		if err != nil {
			c.JSON(http.StatusBadRequest, contracts.ResponseImplementation{
				HttpStatus:       http.StatusBadRequest,
				Explanation:      "invalid definition sent",
				ErrorExplanation: err.Error(),
				Error:            true,
				Success:          false,
			})
		}

		if data != nil {
			kind := ""

			if c.Param("kind") != "" {
				kind = c.Param("kind")
			} else {
				if data["kind"] != nil {
					kind = data["kind"].(string)
				} else {
					c.JSON(http.StatusBadRequest, contracts.ResponseImplementation{
						HttpStatus:       http.StatusBadRequest,
						Explanation:      "",
						ErrorExplanation: "invalid definition sent - kind is not defined",
						Error:            true,
						Success:          false,
					})

					return
				}
			}

			api.ImplementationWrapperApply(authentication.NewUser(c.Request.TLS), kind, jsonData, c)
		} else {
			c.JSON(http.StatusBadRequest, contracts.ResponseImplementation{
				HttpStatus:       http.StatusBadRequest,
				Explanation:      "invalid definition sent",
				ErrorExplanation: err.Error(),
				Error:            true,
				Success:          false,
			})
		}
	}
}

func (api *Api) ImplementationWrapperApply(user *authentication.User, kind string, jsonData []byte, c *gin.Context) {
	var err error
	kindObj, ok := api.KindsRegistry[kind]

	if !ok {
		c.JSON(http.StatusBadRequest, contracts.ResponseImplementation{
			HttpStatus:       http.StatusBadRequest,
			Explanation:      fmt.Sprintf("kind is not present on the server: %s", kind),
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		})

		return
	}

	var response contracts.ResponseImplementation
	response, err = kindObj.Apply(user, jsonData)

	if err != nil {
		c.JSON(http.StatusBadRequest, contracts.ResponseImplementation{
			HttpStatus:       http.StatusInternalServerError,
			Explanation:      err.Error(),
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		})

		return
	}

	c.JSON(response.HttpStatus, response)
	return
}
