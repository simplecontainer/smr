package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/implementations/container/shared"
	"github.com/simplecontainer/smr/pkg/plugins"
	"net/http"
)

func (api *Api) Ps(c *gin.Context) {
	pl := plugins.GetPlugin(api.Config.Root, "container.so")
	containers := pl.GetShared().(*shared.Shared).Registry.Containers

	data, err := json.Marshal(containers)

	if err != nil {

	}

	result := make(map[string]interface{})
	json.Unmarshal(data, &result)

	c.JSON(http.StatusOK, result)
}
