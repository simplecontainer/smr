package api

import (
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/container/shared"
	"github.com/simplecontainer/smr/pkg/network"
	"net/http"
)

func (api *Api) Ps(c *gin.Context) {
	var reg map[string]map[string]platforms.IContainer
	container, ok := api.KindsRegistry["container"]

	if ok {
		reg = container.GetShared().(*shared.Shared).Registry.All()
	}

	if len(reg) > 0 {
		c.JSON(http.StatusOK, contracts.Response{
			Explanation:      "",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
			Data:             network.ToJson(reg),
		})
	} else {
		c.JSON(http.StatusOK, contracts.Response{
			Explanation:      "",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
			Data:             nil,
		})
	}
}
