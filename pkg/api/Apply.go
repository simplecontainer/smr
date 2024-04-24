package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"io"
	"net/http"
	"smr/pkg/implementations"
)

func (api *Api) Apply(c *gin.Context) {
	jsonData, err := io.ReadAll(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid definition sent.",
		})
	} else {
		data := make(map[string]interface{})

		err := json.Unmarshal(jsonData, &data)
		if err != nil {
			panic(err)
		}

		api.ImplementationWrapper(data["kind"].(string), jsonData, c)
	}
}

func (api *Api) ImplementationWrapper(kind string, jsonData []byte, c *gin.Context) {
	plugin, err := getPluginInstance(api.Config.Configuration.Environment.Root, "implementations", kind)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("internal implementation is not present on the server: %s", kind),
			"error":   err.Error(),
			"fail":    true,
			"success": false,
		})

		return
	}

	if plugin != nil {
		ImplementationInternal, err := plugin.Lookup(cases.Title(language.English).String(kind))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("plugin lookup failed: %s", cases.Title(language.English).String(kind)),
				"error":   err.Error(),
				"fail":    true,
				"success": false,
			})

			return
		}

		pl, ok := ImplementationInternal.(implementations.Implementation)

		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "internal implementation malfunctioned on the server",
				"error":   "check server logs",
				"fail":    true,
				"success": false,
			})

			return
		}

		var response implementations.Response
		response, err = pl.Implementation(api.Manager, jsonData)

		c.JSON(response.HttpStatus, gin.H{
			"message": response.Explanation,
			"error":   response.ErrorExplanation,
			"fail":    response.Error,
			"success": response.Success,
		})

		return
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "internal implementation is not present on the server",
			"error":   " internal implementation is not present on the server",
			"fail":    true,
			"success": false,
		})

		return
	}
}
