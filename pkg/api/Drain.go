package api

import (
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	cshared "github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/shared"
	"github.com/simplecontainer/smr/pkg/static"
	"net/http"
	"strconv"
	"time"
)

func (api *Api) Drain(c *gin.Context) {
	c.AddParam("node", strconv.FormatUint(api.Cluster.Node.NodeID, 10))

	api.Manager.KindsRegistry[static.KIND_GITOPS].GetShared().(*shared.Shared).Watchers.Drain()
	api.Manager.KindsRegistry[static.KIND_CONTAINERS].GetShared().(*cshared.Shared).Watchers.Drain()

	go func() {
		for {
			if len(api.Manager.KindsRegistry[static.KIND_GITOPS].GetShared().(*shared.Shared).Watchers.Repositories) == 0 &&
				len(api.Manager.KindsRegistry[static.KIND_CONTAINERS].GetShared().(*cshared.Shared).Watchers.Watchers) == 0 {
				break
			} else {
				time.Sleep(1)
			}
		}

		api.RemoveNode(c)
	}()

	c.JSON(http.StatusOK, common.Response(http.StatusOK, "process of draining the node started", nil, nil))
}
