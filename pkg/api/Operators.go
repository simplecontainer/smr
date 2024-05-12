package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/qdnqn/smr/pkg/operators"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"io"
	"net/http"
	"path/filepath"
)

var invalidOperators []string = []string{
	"Run",
	"ListSupported",
}

func (api *Api) RunOperators(c *gin.Context) {
	group := cleanPath(c.Param("group"))
	operator := c.Param("operator")

	for _, forbidenOperator := range invalidOperators {
		if forbidenOperator == operator {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "this operation is restricted",
				"error":   "can't call internal methods on the operators",
				"fail":    true,
				"success": false,
				"data":    nil,
			})

			return
		}
	}

	plugin, err := getPluginInstance(api.Config.Configuration.Environment.Root, "operators", group)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "operator is not present on the server",
			"error":   err.Error(),
			"fail":    true,
			"success": false,
			"data":    nil,
		})

		return
	}

	if plugin != nil {
		Operator, err := plugin.Lookup(cases.Title(language.English).String(group))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "operator is not present on the server",
				"error":   err.Error(),
				"fail":    true,
				"success": false,
				"data":    nil,
			})

			return
		}

		var pl operators.Operator
		pl, ok := Operator.(operators.Operator)

		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "operator implementation malfunctioned on the server",
				"error":   "check server logs",
				"fail":    true,
				"success": false,
				"data":    nil,
			})

			return
		}

		body := map[string]any{}

		if c.Request.Method == http.MethodPost {
			jsonData, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "invalid JSON sent as the body",
					"error":   err.Error(),
					"fail":    true,
					"success": false,
					"data":    nil,
				})

				return
			}

			err = json.Unmarshal([]byte(jsonData), &body)

			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "invalid JSON sent as the body",
					"error":   err.Error(),
					"fail":    true,
					"success": false,
					"data":    nil,
				})

				return
			}
		}

		request := operators.Request{
			Config:     api.Config,
			Runtime:    api.Runtime,
			Registry:   api.Registry,
			Reconciler: api.Reconciler,
			Manager:    api.Manager,
			Badger:     api.Badger,
			DnsCache:   api.DnsCache,
			Data:       body,
		}

		operatorResponse := pl.Run(operator, request)

		c.JSON(operatorResponse.HttpStatus, gin.H{
			"message": operatorResponse.Explanation,
			"error":   operatorResponse.ErrorExplanation,
			"fail":    operatorResponse.Error,
			"success": operatorResponse.Success,
			"data":    operatorResponse.Data,
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Operator is not present on the server",
		})

		return
	}
}

func (api *Api) ListSupported(c *gin.Context) {
	group := cleanPath(c.Param("group"))

	plugin, err := getPluginInstance(api.Config.Configuration.Environment.Root, "operators", group)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "operator is not present on the server",
			"error":   err.Error(),
			"fail":    true,
			"success": false,
			"data":    nil,
		})

		return
	}

	if plugin != nil {
		Operator, err := plugin.Lookup(cases.Title(language.English).String(group))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "operator is not present on the server",
				"error":   err.Error(),
				"fail":    true,
				"success": false,
				"data":    nil,
			})

			return
		}

		var pl operators.Operator
		pl, ok := Operator.(operators.Operator)

		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "operator implementation malfunctioned on the server",
				"error":   "check server logs",
				"fail":    true,
				"success": false,
				"data":    nil,
			})

			return
		}

		operatorResponse := pl.Run("ListSupported", map[string]any{"test": "test"})
		supportedOperatorList := operatorResponse.Data["SupportedOperations"].([]string)

		for index, supported := range supportedOperatorList {
			for _, forbiddenOperator := range invalidOperators {
				if forbiddenOperator == supported {
					if index+1 > len(supportedOperatorList) {
						supportedOperatorList = supportedOperatorList[:len(supportedOperatorList)-1]
					} else {
						supportedOperatorList = append(supportedOperatorList[:index], supportedOperatorList[index+1:]...)
					}
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "",
			"error":   "",
			"fail":    true,
			"success": false,
			"data":    supportedOperatorList,
		})

		return
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "operator is not present on the server",
			"error":   err.Error(),
			"fail":    true,
			"success": false,
			"data":    nil,
		})

		return
	}
}

func cleanPath(path string) string {
	cleanPath := filepath.Clean(
		path,
	)

	return cleanPath
}
