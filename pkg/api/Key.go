package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/network"
	"io"
	"net/http"
	"strings"
)

// ProposeKey godoc
//
//	@Summary		List kind objects
//	@Description	list kind objects in the store
//	@Tags			database
//	@Produce		json
//
// @Success		200	{object}	  contracts.Response
// @Failure		400	{object}	  contracts.Response
// @Failure		404	{object}	  contracts.Response
// @Failure		500	{object}	  contracts.Response
// @Router		/key/propose/{key} [post]
func (a *Api) ProposeKey(c *gin.Context) {
	data, err := io.ReadAll(c.Request.Body)

	if err != nil {
		common.Response(http.StatusInternalServerError, "", err, nil)
	} else {
		key := c.Param("key")
		a.Cluster.KVStore.Propose(key, data, a.Cluster.Node.NodeID)

		c.JSON(http.StatusOK, common.Response(http.StatusOK, "key stored", nil, nil))
	}
}

// SetKey godoc
//
//	@Summary		List kind objects
//	@Description	list kind objects in the store
//	@Tags			database
//	@Produce		json
//
// @Success		200	{object}	  contracts.Response
// @Failure		400	{object}	  contracts.Response
// @Failure		404	{object}	  contracts.Response
// @Failure		500	{object}	  contracts.Response
// @Router		/key/set/{key} [post]
func (a *Api) SetKey(c *gin.Context) {
	data, err := io.ReadAll(c.Request.Body)

	if err != nil {
		common.Response(http.StatusInternalServerError, "", err, nil)
	} else {
		key := strings.TrimPrefix(c.Param("key"), "/")

		_, err = a.Etcd.Put(context.Background(), key, string(data))

		if err != nil {
			c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
		} else {
			a.Cluster.KVStore.CommittedKeys.Store(key, true)

			c.JSON(http.StatusOK, common.Response(http.StatusOK, "key stored", nil, network.ToJSON(data)))
		}
	}
}

// DeleteKey godoc
//
//	@Summary		List kind objects
//	@Description	list kind objects in the store
//	@Tags			database
//	@Produce		json
//
// @Success		200	{object}	  contracts.Response
// @Failure		400	{object}	  contracts.Response
// @Failure		404	{object}	  contracts.Response
// @Failure		500	{object}	  contracts.Response
// @Router		/key/remove/{key} [delete]
func (a *Api) RemoveKey(c *gin.Context) {
	key := strings.TrimPrefix(c.Param("key"), "/")

	_, err := a.Etcd.Delete(context.Background(), key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
	} else {
		c.JSON(http.StatusOK, common.Response(http.StatusOK, "key deleted", nil, nil))
	}
}
