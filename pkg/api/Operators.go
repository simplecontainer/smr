package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/httpcontract"
	"github.com/simplecontainer/smr/pkg/kinds"
	"io"
	"net/http"
	"path/filepath"
)

var invalidOperators []string = []string{
	"Run",
	"ListSupported",
}

func (api *Api) RunOperators(c *gin.Context) {
	kind := cleanPath(c.Param("group"))
	operator := c.Param("operator")

	for _, forbidenOperator := range invalidOperators {
		if forbidenOperator == operator {
			c.JSON(http.StatusBadRequest, httpcontract.ResponseOperator{
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

	plugin, err := kinds.New(kind)

	if err != nil {
		c.JSON(http.StatusInternalServerError, httpcontract.ResponseOperator{
			HttpStatus:       http.StatusInternalServerError,
			Explanation:      "operator is not present on the server",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})

		return
	}

	body := map[string]any{}

	if c.Request.Method == http.MethodPost {
		var jsonData []byte

		jsonData, err = io.ReadAll(c.Request.Body)

		if err != nil {
			c.JSON(http.StatusBadRequest, httpcontract.ResponseOperator{
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
			c.JSON(http.StatusBadRequest, httpcontract.ResponseOperator{
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

	request := httpcontract.RequestOperator{
		Manager: api.Manager,
		Data:    body,
		User:    authentication.NewUser(c.Request.TLS),
		Client:  api.Manager.Http,
	}

	operatorResponse := plugin.Run(operator, request)

	c.JSON(operatorResponse.HttpStatus, operatorResponse)
}

func (api *Api) ListSupported(c *gin.Context) {
	kind := cleanPath(c.Param("group"))

	plugin, err := kinds.New(kind)

	if err != nil {
		c.JSON(http.StatusBadRequest, httpcontract.ResponseOperator{
			HttpStatus:       http.StatusBadRequest,
			Explanation:      "operator is not present on the server",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})

		return
	}

	request := httpcontract.RequestOperator{
		Manager: api.Manager,
		Data:    nil,
		User:    authentication.NewUser(c.Request.TLS),
		Client:  nil,
	}

	operatorResponse := plugin.Run("ListSupported", request)
	c.JSON(http.StatusOK, operatorResponse)
	return
}

func cleanPath(path string) string {
	cp := filepath.Clean(
		path,
	)

	return cp
}
