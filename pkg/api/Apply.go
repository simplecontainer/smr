package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/qdnqn/smr/pkg/implementations"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"io"
	"net/http"
)

func (api *Api) Apply(c *gin.Context) {
	jsonData, err := io.ReadAll(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusBadRequest, implementations.Response{
			HttpStatus:       http.StatusBadRequest,
			Explanation:      "invalid definition sent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		})
	} else {
		data := make(map[string]interface{})

		err := json.Unmarshal(jsonData, &data)
		if err != nil {
			c.JSON(http.StatusBadRequest, implementations.Response{
				HttpStatus:       http.StatusBadRequest,
				Explanation:      "invalid definition sent",
				ErrorExplanation: err.Error(),
				Error:            true,
				Success:          false,
			})
		}

		api.ImplementationWrapper(data["kind"].(string), jsonData, c)
	}
}

func (api *Api) ImplementationWrapper(kind string, jsonData []byte, c *gin.Context) {
	plugin, err := getPluginInstance(api.Config.Configuration.Environment.Root, "implementations", kind)

	if err != nil {
		c.JSON(http.StatusBadRequest, implementations.Response{
			HttpStatus:       http.StatusBadRequest,
			Explanation:      fmt.Sprintf("internal implementation is not present on the server: %s", kind),
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		})

		return
	}

	if plugin != nil {
		ImplementationInternal, err := plugin.Lookup(cases.Title(language.English).String(kind))
		if err != nil {
			c.JSON(http.StatusBadRequest, implementations.Response{
				HttpStatus:       http.StatusBadRequest,
				Explanation:      fmt.Sprintf("plugin lookup failed: %s", cases.Title(language.English).String(kind)),
				ErrorExplanation: err.Error(),
				Error:            true,
				Success:          false,
			})

			return
		}

		pl, ok := ImplementationInternal.(implementations.Implementation)

		if !ok {
			c.JSON(http.StatusBadRequest, implementations.Response{
				HttpStatus:       http.StatusInternalServerError,
				Explanation:      "internal implementation malfunctioned on the server",
				ErrorExplanation: "",
				Error:            true,
				Success:          false,
			})

			return
		}

		var response implementations.Response
		response, err = pl.Implementation(api.Manager, jsonData)

		if err != nil {
			c.JSON(http.StatusBadRequest, implementations.Response{
				HttpStatus:       http.StatusInternalServerError,
				Explanation:      "internal implementation malfunctioned on the server",
				ErrorExplanation: "",
				Error:            true,
				Success:          false,
			})

			return
		}

		c.JSON(response.HttpStatus, response)
		return
	} else {
		c.JSON(http.StatusBadRequest, implementations.Response{
			HttpStatus:       http.StatusBadRequest,
			Explanation:      fmt.Sprintf("internal implementation is not present on the server: %s", kind),
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		})

		return
	}
}
