package api

import (
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/container/"
	"github.com/simplecontainer/container/container"
	"github.com/simplecontainer/smr/pkg/plugins"
	"net/http"
)

func (api *Api) Ps(c *gin.Context) {
	pl := plugins.GetPlugin(api.Config.Configuration.Environment.Root, "container.so")
	container := pl.

	container.

	data, err := json.Marshal(api.Registry.Containers)

	if err != nil {

	}

	result := make(map[string]interface{})
	json.Unmarshal(data, &result)

	c.JSON(http.StatusOK, result)
}
