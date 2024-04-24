package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"io"
	"net/http"
	"path/filepath"
	"smr/pkg/operators"
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
			c.JSON(http.StatusForbidden, gin.H{
				"message": "Hey there captain. This is not allowed!",
			})

			return
		}
	}

	plugin, err := getPluginInstance(api.Config.Configuration.Environment.Root, "operators", group)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Operator is not present on the server",
		})

		return
	}

	if plugin != nil {
		Operator, err := plugin.Lookup(cases.Title(language.English).String(group))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Operator is not present on the server",
			})

			return
		}

		var pl operators.Operator
		pl, ok := Operator.(operators.Operator)

		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Operator malfunctioned on the server",
			})

			return
		}

		body := map[string]any{}

		if c.Request.Method == http.MethodPost {
			jsonData, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "Invalid JSON sent as the body.",
				})

				return
			}

			err = json.Unmarshal([]byte(jsonData), &body)

			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "Invalid JSON sent as the body.",
				})

				return
			}
		}

		message := pl.Run(operator, body)

		if message["success"] != "true" {
			c.JSON(http.StatusExpectationFailed, message)
		} else {
			c.JSON(http.StatusOK, message)
		}
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
			"message": "Operator is not present on the server",
		})

		return
	}

	if plugin != nil {
		Operator, err := plugin.Lookup(cases.Title(language.English).String(group))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Operator is not present on the server",
			})

			return
		}

		var pl operators.Operator
		pl, ok := Operator.(operators.Operator)

		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Operator malfunctioned on the server",
			})

			return
		}

		message := pl.Run("ListSupported", map[string]any{"test": "test"})
		supportedOperatorList := message["SupportedOperations"].([]string)

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

		c.JSON(http.StatusOK, gin.H{"SupportedOperators": supportedOperatorList})

		return
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Operator is not present on the server",
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
