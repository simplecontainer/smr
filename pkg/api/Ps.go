package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (api *Api) Ps(c *gin.Context) {
	data, err := json.Marshal(api.Registry.Containers)

	if err != nil {

	}

	result := make(map[string]interface{})
	json.Unmarshal(data, &result)

	c.JSON(http.StatusOK, result)
}
