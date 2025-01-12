package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts"
	"io"
	"net/http"
	"path/filepath"
)

var invalidOperators []string = []string{
	"Run",
	"ListSupported",
}

func (api *Api) RunControl(c *gin.Context) {
	kind := cleanPath(c.Param("kind"))
	operation := c.Param("operation")
	group := c.Param("group")
	name := c.Param("name")

	for _, forbidenOperator := range invalidOperators {
		if forbidenOperator == operation {
			c.JSON(http.StatusBadRequest, contracts.Response{
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

	var err error
	kindObj, ok := api.KindsRegistry[kind]

	if !ok {
		c.JSON(http.StatusInternalServerError, contracts.Response{
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
			c.JSON(http.StatusBadRequest, contracts.Response{
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
			c.JSON(http.StatusBadRequest, contracts.Response{
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

	operatorResponse := kindObj.Run(operation, contracts.Control{
		Kind:      kind,
		Operation: operation,
		Group:     group,
		Name:      name,
		Data:      body,
		User:      authentication.NewUser(c.Request.TLS),
	})

	c.JSON(operatorResponse.HttpStatus, operatorResponse)
}

func (api *Api) ListSupported(c *gin.Context) {
	kind := cleanPath(c.Param("kind"))

	kindObj, ok := api.KindsRegistry[kind]

	if !ok {
		c.JSON(http.StatusBadRequest, contracts.Response{
			HttpStatus:       http.StatusBadRequest,
			Explanation:      "operator is not present on the server",
			ErrorExplanation: "operator is not present on the server",
			Error:            true,
			Success:          false,
			Data:             nil,
		})

		return
	}

	request := contracts.Control{
		Data: nil,
		User: authentication.NewUser(c.Request.TLS),
	}

	operatorResponse := kindObj.Run("ListSupported", request)
	c.JSON(http.StatusOK, operatorResponse)

	return
}

func cleanPath(path string) string {
	cp := filepath.Clean(
		path,
	)

	return cp
}
