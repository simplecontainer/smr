package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/implementations/container/shared"
	"github.com/simplecontainer/smr/pkg/plugins"
	"net/http"
)

func (api *Api) Ps(c *gin.Context) {
	pl := plugins.GetPlugin(api.Config.OptRoot, "container.so")
	registry := pl.GetShared().(*shared.Shared).Registry

	if registry != nil {
		if len(registry.Containers) > 0 {
			data, err := json.Marshal(registry.Containers)
			result := make(map[string]interface{})

			if err != nil {
				c.JSON(http.StatusInternalServerError, result)
				return
			}

			json.Unmarshal(data, &result)
			c.JSON(http.StatusOK, result)
		} else {
			c.JSON(http.StatusOK, "{}")
		}
	} else {
		result := make(map[string]interface{})
		c.JSON(http.StatusOK, result)
	}
}
