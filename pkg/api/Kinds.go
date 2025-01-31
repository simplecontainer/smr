package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/network"
	clientv3 "go.etcd.io/etcd/client/v3"
	"io"
	"net/http"
)

// ListKind godoc
//
//	@Summary		List kind objects
//	@Description	list kind objects in the store
//	@Tags			database
//	@Produce		json

// @Success		200	{object}	  contracts.Response
// @Failure		400	{object}	  contracts.Response
// @Failure		404	{object}	  contracts.Response
// @Failure		500	{object}	  contracts.Response
// @Router			/kind/{prefix}/{category}/{kind} [get]
func (api *Api) ListKind(c *gin.Context) {
	prefix := c.Param("prefix")
	category := c.Param("category")
	kind := c.Param("kind")

	response, err := api.Etcd.Get(context.Background(), fmt.Sprintf("/%s/%s/%s", prefix, category, kind), clientv3.WithPrefix())

	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
	} else {
		c.JSON(http.StatusOK, common.Response(http.StatusOK, "", nil, network.ToJson(response.Kvs)))
	}
}

// ListKindGroup godoc
//
//	@Summary		List kind objects for group
//	@Description	list kind objects in the store for specific group
//	@Tags			database
//	@Produce		json

// @Success		200	{object}	  contracts.Response
// @Failure		400	{object}	  contracts.Response
// @Failure		404	{object}	  contracts.Response
// @Failure		500	{object}	  contracts.Response
// @Router		/kind/{prefix}/{category}/{kind}/{group} [get]
func (api *Api) ListKindGroup(c *gin.Context) {
	prefix := c.Param("prefix")
	category := c.Param("category")
	kind := c.Param("kind")
	group := c.Param("group")

	response, err := api.Etcd.Get(context.Background(), fmt.Sprintf("/%s/%s/%s/%s", prefix, category, kind, group), clientv3.WithPrefix())

	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
	} else {
		c.JSON(http.StatusOK, common.Response(http.StatusInternalServerError, "", nil, network.ToJson(response.Kvs)))
	}
}

// GetKind godoc
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
func (api *Api) GetKind(c *gin.Context) {
	prefix := c.Param("prefix")
	category := c.Param("category")
	kind := c.Param("kind")
	group := c.Param("group")
	name := c.Param("name")

	response, err := api.Etcd.Get(context.Background(), fmt.Sprintf("/%s/%s/%s/%s/%s", prefix, category, kind, group, name), clientv3.WithPrefix())

	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
	} else {
		if len(response.Kvs) == 0 {
			c.JSON(http.StatusNotFound, common.Response(http.StatusNotFound, "", errors.New("key not found"), nil))
		} else {
			var bytes []byte
			bytes, err = json.RawMessage(response.Kvs[0].Value).MarshalJSON()

			c.JSON(http.StatusOK, common.Response(http.StatusOK, "", nil, bytes))
		}
	}
}

// SetKind godoc
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
// @Router		/kind/propose/{prefix}/{category}/{kind}/{group}/{name} [post]
func (api *Api) ProposeKind(c *gin.Context) {
	data, err := io.ReadAll(c.Request.Body)

	if err != nil {
		common.Response(http.StatusInternalServerError, "", err, nil)
	} else {
		prefix := c.Param("prefix")
		category := c.Param("category")
		kind := c.Param("kind")
		group := c.Param("group")
		name := c.Param("name")

		format, _ := f.New(prefix, category, kind, group, name)
		api.Cluster.KVStore.Propose(format.ToStringWithUUID(), data, api.Cluster.Node.NodeID)

		c.JSON(http.StatusOK, common.Response(http.StatusOK, "object stored", nil, nil))
	}
}

// SetKind godoc
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
// @Router		/kind/{prefix}/{category}/{kind}/{group}/{name} [post]
func (api *Api) SetKind(c *gin.Context) {
	data, err := io.ReadAll(c.Request.Body)

	if err != nil {
		common.Response(http.StatusInternalServerError, "", err, nil)
	} else {
		prefix := c.Param("prefix")
		category := c.Param("category")
		kind := c.Param("kind")
		group := c.Param("group")
		name := c.Param("name")

		_, err = api.Etcd.Put(context.Background(), fmt.Sprintf("/%s/%s/%s/%s/%s", prefix, category, kind, group, name), string(data))

		if err != nil {
			c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
		} else {
			c.JSON(http.StatusOK, common.Response(http.StatusOK, "object stored", nil, network.ToJson(data)))
		}
	}
}

// DeleteKind godoc
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
// @Router		/kind/{prefix}/{category}/{kind}/{group}/{name} [delete]
func (api *Api) DeleteKind(c *gin.Context) {
	prefix := c.Param("prefix")
	category := c.Param("category")
	kind := c.Param("kind")
	group := c.Param("group")
	name := c.Param("name")

	_, err := api.Etcd.Delete(context.Background(), fmt.Sprintf("/%s/%s/%s/%s/%s", prefix, category, kind, group, name))

	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
	} else {
		c.JSON(http.StatusOK, common.Response(http.StatusOK, "object deleted", nil, nil))
	}
}
