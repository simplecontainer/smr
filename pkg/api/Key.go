package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/network"
	"io"
	"net/http"
	"strings"
)

// GetKey godoc
//
//	@Summary		Get specific kind
//	@Description	get specific kind from the store
//	@Tags			database
//	@Produce		json
//
// @Success		200	{object}	  contracts.Response
// @Failure		400	{object}	  contracts.Response
// @Failure		404	{object}	  contracts.Response
// @Failure		500	{object}	  contracts.Response
// @Router		/kind/{prefix}/{category}/{kind}/{group}/{name} [get]
func (api *Api) GetKey(c *gin.Context) {
	key := strings.TrimPrefix(c.Param("key"), "/")

	response, err := api.Etcd.Get(context.Background(), key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
	} else {
		if len(response.Kvs) == 0 {
			c.JSON(http.StatusNotFound, common.Response(http.StatusNotFound, "", errors.New("key not found"), nil))
		} else {
			var bytes json.RawMessage
			bytes, err = json.RawMessage(response.Kvs[0].Value).MarshalJSON()

			c.JSON(http.StatusOK, common.Response(http.StatusOK, "", nil, bytes))
		}
	}
}

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
func (api *Api) ProposeKey(c *gin.Context) {
	data, err := io.ReadAll(c.Request.Body)

	if err != nil {
		common.Response(http.StatusInternalServerError, "", err, nil)
	} else {
		key := c.Param("key")
		api.Cluster.KVStore.Propose(key, data, api.Cluster.Node.NodeID)

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
// @Router		/key/{key} [post]
func (api *Api) SetKey(c *gin.Context) {
	data, err := io.ReadAll(c.Request.Body)

	if err != nil {
		common.Response(http.StatusInternalServerError, "", err, nil)
	} else {
		key := strings.TrimPrefix(c.Param("key"), "/")

		_, err = api.Etcd.Put(context.Background(), key, string(data))

		if err != nil {
			c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
		} else {
			c.JSON(http.StatusOK, common.Response(http.StatusOK, "key stored", nil, network.ToJson(data)))
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
// @Router		/key/{key} [delete]
func (api *Api) DeleteKey(c *gin.Context) {
	key := strings.TrimPrefix(c.Param("key"), "/")

	_, err := api.Etcd.Delete(context.Background(), key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
	} else {
		c.JSON(http.StatusOK, common.Response(http.StatusOK, "key deleted", nil, nil))
	}
}
