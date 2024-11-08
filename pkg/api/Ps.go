package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/kinds/container/registry"
	"github.com/simplecontainer/smr/pkg/kinds/container/shared"
	"net/http"
)

func (api *Api) Ps(c *gin.Context) {
	var reg *registry.Registry
	container, ok := api.KindsRegistry["container"]

	if ok {
		reg = container.GetShared().(*shared.Shared).Registry
	}

	if reg != nil {
		data, err := json.Marshal(reg)
		result := make(map[string]interface{})

		if err != nil {
			c.JSON(http.StatusInternalServerError, result)
			return
		}

		err = json.Unmarshal(data, &result)

		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
		}

		c.JSON(http.StatusOK, result)
	} else {
		result := make(map[string]interface{})
		c.JSON(http.StatusOK, result)
	}
}
