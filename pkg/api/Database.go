package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/helpers"
	"github.com/simplecontainer/smr/pkg/network"
	clientv3 "go.etcd.io/etcd/client/v3"
	"io"
	"net/http"
	"strings"
)

// DatabaseGet godoc
//
//	@Summary		Get value from the key-value store
//	@Description	get string by key from the key-value store
//	@Tags			database
//	@Produce		json
//	@Param			key	path		string	true	"RandomKey"
//	@Success		200	{object}	  contracts.Response
//	@Failure		400	{object}	  contracts.Response
//	@Failure		404	{object}	  contracts.Response
//	@Failure		500	{object}	  contracts.Response
//	@Router			/database/{key} [get]
func (api *Api) DatabaseGet(c *gin.Context) {
	key := strings.TrimPrefix(c.Param("key"), "/")
	response, err := api.Etcd.Get(context.Background(), key)

	if err != nil {
		c.JSON(http.StatusNotFound, contracts.Response{
			Explanation:      "failed to read from the key-value store",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})
	} else {
		if response.Count == 0 {
			c.JSON(http.StatusNotFound, contracts.Response{
				Explanation:      "failed to read from the key-value store",
				ErrorExplanation: errors.New("key not found").Error(),
				Error:            true,
				Success:          false,
				Data:             nil,
			})
		} else {
			c.JSON(http.StatusOK, contracts.Response{
				Explanation:      "found key in the key-value store",
				ErrorExplanation: "",
				Error:            false,
				Success:          true,
				Data:             response.Kvs[0].Value,
			})
		}
	}
}

// DatabaseSet godoc
//
//	@Summary		Set value in the key-value store
//	@Description	set string by key in the key-value store
//	@Tags			database
//	@Accepts		json
//	@Produce		json
//	@Param			key		path		string	true	"RandomKey"
//	@Param			value	body		Kv		true	"value"
//	@Success		200		{object}	  contracts.Response
//	@Failure		400		{object}	  contracts.Response
//	@Failure		404		{object}	  contracts.Response
//	@Failure		500		{object}	  contracts.Response
//	@Router			/database/{key} [post]
func (api *Api) DatabaseSet(c *gin.Context) {
	var data []byte
	var err error

	data, err = io.ReadAll(c.Request.Body)

	key := strings.TrimPrefix(c.Param("key"), "/")

	if err == nil {
		_, err = api.Etcd.Put(context.Background(), key, string(data))

		if err != nil {
			c.JSON(http.StatusInternalServerError, contracts.Response{
				Explanation:      "failed to store value in the key-value store",
				ErrorExplanation: err.Error(),
				Error:            true,
				Success:          false,
				Data:             nil,
			})
		} else {
			c.JSON(http.StatusOK, contracts.Response{
				Explanation:      "value stored in the key value store",
				ErrorExplanation: "",
				Error:            false,
				Success:          true,
				Data:             network.ToJson(data),
			})
		}
	} else {
		c.JSON(http.StatusInternalServerError, contracts.Response{
			Explanation:      "failed to store value in the key-value store",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})
	}
}

func (api *Api) ProposeDatabase(c *gin.Context) {
	data, err := io.ReadAll(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusInternalServerError, contracts.Response{
			Explanation:      "failed to store value in the key-value store",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})

		return
	}

	if api.Cluster.KVStore == nil {
		c.JSON(http.StatusInternalServerError, contracts.Response{
			Explanation:      "key-value store is not started yet",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})

		return
	}

	key := strings.TrimPrefix(c.Param("key"), "/")

	format := f.NewFromString(key)
	api.Cluster.KVStore.Propose(format.ToStringWithUUID(), data, helpers.Category(c.Param("type")), api.Cluster.Node.NodeID)

	// To prevent empty responses since Json.RawMessage is in the response
	if len(data) == 0 {
		data, _ = json.Marshal("{}")
	}

	c.JSON(http.StatusOK, contracts.Response{
		Explanation:      "value proposed to the key value store",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             network.ToJson(data),
	})
}

// DatabaseGetKeysPrefix godoc
//
//	@Summary		Get keys by prefix in the key-value store
//	@Description	get all keys by prefix in the key-value store
//	@Tags			database
//	@Produce		json
//	@Success		200	{object}	  contracts.Response
//	@Failure		400	{object}	  contracts.Response
//	@Failure		404	{object}	  contracts.Response
//	@Failure		500	{object}	  contracts.Response
//	@Router			/database/{key}/{prefix} [get]
func (api *Api) DatabaseGetKeysPrefix(c *gin.Context) {
	prefix := []byte(strings.TrimPrefix(c.Param("prefix"), "/"))

	response, err := api.Etcd.Get(context.Background(), string(prefix), clientv3.WithPrefix())

	var keys []string

	for _, kv := range response.Kvs {
		keys = append(keys, string(kv.Key))
	}

	if err == nil {
		c.JSON(http.StatusOK, contracts.Response{
			HttpStatus:       http.StatusOK,
			Explanation:      "keys found",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
			Data:             network.ToJson(keys),
		})
	} else {
		c.JSON(http.StatusNotFound, contracts.Response{
			HttpStatus:       http.StatusNotFound,
			Explanation:      "failed to retrieve keys from the key-value store",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})
	}
}

// DatabaseGetKeys godoc
//
//	@Summary		Get keys by prefix in the key-value store
//	@Description	get all keys by prefix in the key-value store
//	@Tags			database
//	@Produce		json
//	@Success		200	{object}	  contracts.Response
//	@Failure		400	{object}	  contracts.Response
//	@Failure		404	{object}	  contracts.Response
//	@Failure		500	{object}	  contracts.Response
//	@Router			/database/keys [get]
func (api *Api) DatabaseGetKeys(c *gin.Context) {
	response, err := api.Etcd.Get(context.Background(), "", clientv3.WithPrefix())

	var keys []string

	for _, kv := range response.Kvs {
		keys = append(keys, string(kv.Key))
	}

	if err == nil {
		c.JSON(http.StatusOK, contracts.Response{
			Explanation:      "succesfully retrieved keys from the key-value store",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
			Data:             network.ToJson(keys),
		})
	} else {
		c.JSON(http.StatusNotFound, contracts.Response{
			Explanation:      "failed to retrieve keys from the key-value store",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})
	}
}

// DatabaseRemoveKeys godoc
//
//	@Summary		Remove keys by prefix in the key-value store
//	@Description	remove all keys by prefix in the key-value store
//	@Tags			database
//	@Produce		json
//	@Success		200	{object}	  contracts.Response
//	@Failure		400	{object}	  contracts.Response
//	@Failure		404	{object}	  contracts.Response
//	@Failure		500	{object}	  contracts.Response
//	@Router			/database/keys [delete]
func (api *Api) DatabaseRemoveKeys(c *gin.Context) {
	var keys []string

	prefix := []byte(strings.TrimPrefix(c.Param("prefix"), "/"))

	response, err := api.Etcd.Delete(context.Background(), string(prefix), clientv3.WithPrefix())
	if err != nil {
		return
	}

	for _, kv := range response.PrevKvs {
		keys = append(keys, string(kv.Key))
	}

	if err == nil {
		c.JSON(http.StatusOK, contracts.Response{
			Explanation:      "succesfully removed keys from the key-value store",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
			Data:             network.ToJson(keys),
		})
	} else {
		c.JSON(http.StatusNotFound, contracts.Response{
			Explanation:      "failed to remove keys from the key-value store",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             network.ToJson(keys),
		})
	}
}
