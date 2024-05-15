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
			c.JSON(http.StatusBadRequest, operators.Response{
				HttpStatus:       http.StatusBadRequest,
				Explanation:      "this operation is restricted",
				ErrorExplanation: "can't call internal methods on the operators",
				Error:            true,
				Success:          false,
				Data:             nil,
			})

			return
		}
	}

	plugin, err := getPluginInstance(api.Config.Configuration.Environment.Root, "operators", group)

	if err != nil {
		c.JSON(http.StatusInternalServerError, operators.Response{
			HttpStatus:       http.StatusInternalServerError,
			Explanation:      "operator is not present on the server",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})

		return
	}

	if plugin != nil {
		Operator, err := plugin.Lookup(cases.Title(language.English).String(group))
		if err != nil {
			c.JSON(http.StatusInternalServerError, operators.Response{
				HttpStatus:       http.StatusInternalServerError,
				Explanation:      "operator is not present on the server",
				ErrorExplanation: err.Error(),
				Error:            true,
				Success:          false,
				Data:             nil,
			})

			return
		}

		var pl operators.Operator
		pl, ok := Operator.(operators.Operator)

		if !ok {
			c.JSON(http.StatusInternalServerError, operators.Response{
				HttpStatus:       http.StatusInternalServerError,
				Explanation:      "operator implementation malfunctioned on the server",
				ErrorExplanation: "check server logs",
				Error:            true,
				Success:          false,
				Data:             nil,
			})

			return
		}

		body := map[string]any{}

		if c.Request.Method == http.MethodPost {
			jsonData, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusBadRequest, operators.Response{
					HttpStatus:       http.StatusBadRequest,
					Explanation:      "invalid JSON sent as the body",
					ErrorExplanation: err.Error(),
					Error:            true,
					Success:          false,
					Data:             nil,
				})

				return
			}

			err = json.Unmarshal([]byte(jsonData), &body)

			if err != nil {
				c.JSON(http.StatusBadRequest, operators.Response{
					HttpStatus:       http.StatusBadRequest,
					Explanation:      "invalid JSON sent as the body",
					ErrorExplanation: err.Error(),
					Error:            true,
					Success:          false,
					Data:             nil,
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

		c.JSON(operatorResponse.HttpStatus, operatorResponse)
	} else {
		c.JSON(http.StatusInternalServerError, operators.Response{
			HttpStatus:       http.StatusInternalServerError,
			Explanation:      "operator is not present on the server",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})

		return
	}
}

func (api *Api) ListSupported(c *gin.Context) {
	group := cleanPath(c.Param("group"))

	plugin, err := getPluginInstance(api.Config.Configuration.Environment.Root, "operators", group)

	if err != nil {
		c.JSON(http.StatusBadRequest, operators.Response{
			HttpStatus:       http.StatusBadRequest,
			Explanation:      "operator is not present on the server",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})

		return
	}

	if plugin != nil {
		Operator, err := plugin.Lookup(cases.Title(language.English).String(group))
		if err != nil {
			c.JSON(http.StatusBadRequest, operators.Response{
				HttpStatus:       http.StatusBadRequest,
				Explanation:      "operator is not present on the server",
				ErrorExplanation: err.Error(),
				Error:            true,
				Success:          false,
				Data:             nil,
			})

			return
		}

		var pl operators.Operator
		pl, ok := Operator.(operators.Operator)

		if !ok {
			c.JSON(http.StatusInternalServerError, operators.Response{
				HttpStatus:       http.StatusInternalServerError,
				Explanation:      "operator implementation malfunctioned on the server",
				ErrorExplanation: err.Error(),
				Error:            true,
				Success:          false,
				Data:             nil,
			})

			return
		}

		operatorResponse := pl.Run("ListSupported", map[string]any{})
		c.JSON(http.StatusOK, operatorResponse)
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
