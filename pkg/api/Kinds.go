package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/metrics"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/wI2L/jsondiff"
	clientv3 "go.etcd.io/etcd/client/v3"
	"io"
	"net/http"
)

// List godoc
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
// @Router			/kind/{prefix}/{category}/{kind} [get]
func (a *Api) List(c *gin.Context) {
	response, err := a.Etcd.Get(c.Request.Context(), fmt.Sprintf("/"), clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))

	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
	} else {
		kinds := make([]string, 0)

		for _, kv := range response.Kvs {
			kinds = append(kinds, string(kv.Key))
		}

		c.JSON(http.StatusOK, common.Response(http.StatusOK, "", nil, network.ToJSON(kinds)))
	}
}

// ListKind godoc
//
//	@Summary		List kind objects
//	@Description	list kind objects in the store
//	@Tags			database
//	@Produce		json
func (a *Api) ListKind(c *gin.Context) {
	prefix := c.Param("prefix")
	version := c.Param("version")
	category := c.Param("category")
	kind := c.Param("kind")
	group := c.Param("group")

	format := f.New(prefix, version, category, kind, group)
	opts := f.DefaultToStringOpts()
	opts.AddPrefixSlash = true
	opts.AddTrailingSlash = true
	response, err := a.Etcd.Get(c.Request.Context(), format.ToStringWithOpts(opts), clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	send(c, response, err, nil)
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
func (a *Api) GetKind(c *gin.Context) {
	metrics.DatabaseGet.Increment()
	prefix := c.Param("prefix")
	version := c.Param("version")
	category := c.Param("category")
	kind := c.Param("kind")
	group := c.Param("group")
	name := c.Param("name")
	field := c.Param("field")

	format := f.New(prefix, version, category, kind, group, name, field)
	opts := f.DefaultToStringOpts()
	opts.AddPrefixSlash = true
	response, err := a.Etcd.Get(c.Request.Context(), format.ToStringWithOpts(opts), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))

	if err != nil || len(response.Kvs) == 0 {
		c.JSON(http.StatusNotFound, common.Response(http.StatusNotFound, "", errors.New("resource not found"), nil))
		return
	}

	var bytes json.RawMessage
	bytes, err = json.RawMessage(response.Kvs[0].Value).MarshalJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
		return
	}

	c.JSON(http.StatusOK, common.Response(http.StatusOK, "", nil, bytes))
}

// ProposeKind godoc
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
func (a *Api) ProposeKind(c *gin.Context) {
	metrics.DatabasePropose.Increment()
	data, err := io.ReadAll(c.Request.Body)

	if err != nil {
		common.Response(http.StatusInternalServerError, "", err, nil)
		return
	}

	prefix := c.Param("prefix")
	version := c.Param("version")
	category := c.Param("category")
	kind := c.Param("kind")
	group := c.Param("group")
	name := c.Param("name")
	field := c.Param("field")

	format := f.New(prefix, version, category, kind, group, name, field)
	a.Cluster.KVStore.Propose(format.ToStringWithUUID(), data, a.Cluster.Node.NodeID)

	c.JSON(http.StatusOK, common.Response(http.StatusOK, format.ToString(), nil, nil))
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
func (a *Api) SetKind(c *gin.Context) {
	metrics.DatabaseSet.Increment()
	data, err := io.ReadAll(c.Request.Body)

	if err != nil {
		common.Response(http.StatusInternalServerError, "", err, nil)
		return
	}

	prefix := c.Param("prefix")
	version := c.Param("version")
	category := c.Param("category")
	kind := c.Param("kind")
	group := c.Param("group")
	name := c.Param("name")
	field := c.Param("field")

	format := f.New(prefix, version, category, kind, group, name, field)
	opts := f.DefaultToStringOpts()
	opts.AddPrefixSlash = true
	_, err = a.Etcd.Put(c.Request.Context(), format.ToStringWithOpts(opts), string(data))

	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
		return
	}

	a.Cluster.KVStore.CommittedKeys.Store(format.ToStringWithOpts(opts), true)
	c.JSON(http.StatusOK, common.Response(http.StatusOK, "object stored", nil, network.ToJSON(data)))
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
func (a *Api) DeleteKind(c *gin.Context) {
	metrics.DatabaseRemove.Increment()

	prefix := c.Param("prefix")
	version := c.Param("version")
	category := c.Param("category")
	kind := c.Param("kind")
	group := c.Param("group")
	name := c.Param("name")
	field := c.Param("field")

	format := f.New(prefix, version, category, kind, group, name, field)
	opts := f.DefaultToStringOpts()
	opts.AddPrefixSlash = true
	_, err := a.Etcd.Delete(c.Request.Context(), format.ToStringWithOpts(opts))

	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
		return
	}

	c.JSON(http.StatusOK, common.Response(http.StatusOK, "object deleted", nil, nil))
}

// CompareKind godoc
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
func (a *Api) CompareKind(c *gin.Context) {
	prefix := c.Param("prefix")
	version := c.Param("version")
	category := c.Param("category")
	kind := c.Param("kind")
	group := c.Param("group")
	name := c.Param("name")
	field := c.Param("field")

	format := f.New(prefix, version, category, kind, group, name, field)
	opts := f.DefaultToStringOpts()
	opts.AddPrefixSlash = true
	response, err := a.Etcd.Get(c.Request.Context(), format.ToStringWithOpts(opts), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))

	if err != nil || len(response.Kvs) == 0 {
		c.JSON(http.StatusNotFound, common.Response(http.StatusNotFound, "", errors.New("resource not found"), nil))
		return
	}

	data, err := io.ReadAll(c.Request.Body)

	if err != nil {
		common.Response(http.StatusInternalServerError, "", err, nil)
		return
	}

	changelog, _ := jsondiff.CompareJSON(data, response.Kvs[0].Value)

	var bytes []byte
	bytes, err = json.Marshal(changelog)

	if err != nil {
		common.Response(http.StatusInternalServerError, "", err, nil)
		return
	}

	c.JSON(http.StatusOK, common.Response(http.StatusOK, "", nil, bytes))
}

func send(c *gin.Context, response *clientv3.GetResponse, err error, additionalData interface{}) {
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
		return
	}

	kinds := make([]json.RawMessage, 0)
	for _, kv := range response.Kvs {
		kinds = append(kinds, kv.Value)
	}

	c.JSON(http.StatusOK, common.Response(http.StatusOK, "", nil, network.ToJSON(kinds)))
}
