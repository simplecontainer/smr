package api

import (
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/static"
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
	api.BadgerSync.RLock()

	key := strings.TrimPrefix(c.Param("key"), "/")

	var value []byte

	err := api.Badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		var valueCopy []byte
		valueCopy, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}

		value = valueCopy

		return nil
	})

	api.BadgerSync.RUnlock()

	if err != nil {
		c.JSON(http.StatusNotFound, contracts.Response{
			Explanation:      "failed to read from the key-value store",
			ErrorExplanation: err.Error(),
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
			Data:             value,
		})
	}
}

// DatabaseGetBase64 godoc
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
func (api *Api) DatabaseGetBase64(c *gin.Context) {
	api.BadgerSync.RLock()

	key := strings.TrimPrefix(c.Param("key"), "/")
	var value []byte

	err := api.Badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		var valueCopy []byte
		valueCopy, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}

		value = valueCopy

		return nil
	})

	api.BadgerSync.RUnlock()

	if err != nil {
		c.JSON(http.StatusNotFound, contracts.Response{
			Explanation:      "failed to read from the key-value store",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})
	} else {
		bytes, _ := json.Marshal(value)

		c.JSON(http.StatusOK, contracts.Response{
			Explanation:      "found key in the key-value store",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
			Data:             bytes,
		})
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
		api.BadgerSync.Lock()

		err = api.Badger.Update(func(txn *badger.Txn) error {
			err = txn.Set([]byte(key), data)
			return err
		})

		api.BadgerSync.Unlock()

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
				Data:             data,
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

func (api *Api) DatabaseSetBase64(c *gin.Context) {
	var data []byte
	var err error

	data, err = io.ReadAll(c.Request.Body)

	key := strings.TrimPrefix(c.Param("key"), "/")

	if err == nil {
		api.BadgerSync.Lock()

		err = api.Badger.Update(func(txn *badger.Txn) error {
			err = txn.Set([]byte(key), data)
			return err
		})

		api.BadgerSync.Unlock()

		if err != nil {
			c.JSON(http.StatusInternalServerError, contracts.Response{
				Explanation:      "failed to store value in the key-value store",
				ErrorExplanation: err.Error(),
				Error:            true,
				Success:          false,
				Data:             nil,
			})
		} else {
			bytes, _ := json.Marshal(data)

			c.JSON(http.StatusOK, contracts.Response{
				Explanation:      "value stored in the key value store",
				ErrorExplanation: "",
				Error:            false,
				Success:          true,
				Data:             bytes,
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

func (api *Api) Propose(c *gin.Context) {
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

	// To prevent empty responses since Json.RawMessage is in the response
	if len(data) == 0 {
		data, _ = json.Marshal("{}")
	}

	switch c.Param("type") {
	case static.CATEGORY_PLAIN:
		api.Cluster.KVStore.Propose(key, data, api.Config.Node)
		c.JSON(http.StatusOK, contracts.Response{
			Explanation:      "value stored in the key value store",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
			Data:             data,
		})
		return
	case static.CATEGORY_OBJECT:
		fmt.Println("PROPOSE")

		api.Cluster.KVStore.ProposeObject(key, data, api.Config.Node)
		c.JSON(http.StatusOK, contracts.Response{
			Explanation:      "value stored in the key value store",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
			Data:             data,
		})
		return
	case static.CATEGORY_SECRET:
		api.Cluster.KVStore.ProposeSecret(key, data, api.Config.Node)

		bytes, _ := json.Marshal(data)

		c.JSON(http.StatusOK, contracts.Response{
			Explanation:      "value stored in the key value store",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
			Data:             bytes,
		})
		return
	case static.CATEGORY_ETCD:
		api.Cluster.KVStore.ProposeEtcd(key, data, api.Config.Node)
		c.JSON(http.StatusOK, contracts.Response{
			Explanation:      "value stored in the key value store",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
			Data:             data,
		})
		return
	}

	c.JSON(http.StatusBadRequest, contracts.Response{
		Explanation:      "",
		ErrorExplanation: "invalid category selected for the propose",
		Error:            true,
		Success:          false,
		Data:             nil,
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
	var keys []string

	prefix := []byte(strings.TrimPrefix(c.Param("prefix"), "/"))

	api.BadgerSync.RLock()

	err := api.Badger.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()

			keys = append(keys, string(k))
		}

		return nil
	})

	api.BadgerSync.RUnlock()

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
	var keys []string

	api.BadgerSync.RLock()

	err := api.Badger.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()

			keys = append(keys, string(k))
		}

		return nil
	})

	api.BadgerSync.RUnlock()

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

	api.BadgerSync.Lock()

	err := api.Badger.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false

		it := txn.NewIterator(opts)
		defer it.Close()

		var err error

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			err = txn.Delete(it.Item().KeyCopy(nil))

			if err != nil {
				logger.Log.Error(err.Error())
				return err
			}
		}

		return err
	})

	api.BadgerSync.Unlock()

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
