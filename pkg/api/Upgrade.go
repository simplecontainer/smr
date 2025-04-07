package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	cshared "github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/shared"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/simplecontainer/smr/pkg/upgrader"
	"io"
	"net/http"
	"strconv"
	"time"
)

func (api *Api) Upgrade(c *gin.Context) {
	data, err := io.ReadAll(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "failed to start upgrade - try again", err, nil))

	}

	var u *upgrader.Upgrade

	err = json.Unmarshal(data, &u)

	if err != nil {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "invalid node sent", err, nil))
	}

	c.AddParam("node", strconv.FormatUint(api.Cluster.Node.NodeID, 10))

	api.Cluster.Node.SetDrain(true)
	api.Cluster.Node.SetUpgrade(true)

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

		upgrade := upgrader.New(u.Image, u.Tag)
		err = upgrade.Apply(c, api.Etcd)

		if err != nil {
			c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "failed to start upgrade - try again", err, nil))
		}
	}()

	c.JSON(http.StatusOK, common.Response(http.StatusOK, "process of draining the node started", nil, nil))
}
