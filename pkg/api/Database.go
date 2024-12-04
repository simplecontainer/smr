package api

import (
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/logger"
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
//	@Success		200	{object}	  contracts.ResponseOperator
//	@Failure		400	{object}	  contracts.ResponseOperator
//	@Failure		404	{object}	  contracts.ResponseOperator
//	@Failure		500	{object}	  contracts.ResponseOperator
//	@Router			/database/{key} [get]
func (api *Api) DatabaseGet(c *gin.Context) {
	api.BadgerSync.RLock()

	key := strings.TrimPrefix(c.Param("key"), "/")

	err := api.Badger.View(func(txn *badger.Txn) error {
		var value []byte

		item, err := txn.Get([]byte(key))
		if err != nil {
			c.JSON(http.StatusNotFound, contracts.ResponseOperator{
				Explanation:      "key not found",
				ErrorExplanation: "",
				Error:            true,
				Success:          false,
				Data:             nil,
			})

			return nil
		}

		value, err = item.ValueCopy(nil)
		if err != nil {
			c.JSON(http.StatusNotFound, contracts.ResponseOperator{
				Explanation:      "key not found",
				ErrorExplanation: "",
				Error:            true,
				Success:          false,
				Data:             nil,
			})

			return nil
		}

		c.JSON(http.StatusOK, contracts.ResponseOperator{
			Explanation:      "found key in the key-value store",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
			Data: map[string]any{
				key: value,
			},
		})

		return nil
	})

	api.BadgerSync.RUnlock()

	if err != nil {
		logger.Log.Error(err.Error())

		c.JSON(http.StatusNotFound, contracts.ResponseOperator{
			Explanation:      "failed to read from the key-value store",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
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
//	@Success		200		{object}	  contracts.ResponseOperator
//	@Failure		400		{object}	  contracts.ResponseOperator
//	@Failure		404		{object}	  contracts.ResponseOperator
//	@Failure		500		{object}	  contracts.ResponseOperator
//	@Router			/database/{key} [post]
func (api *Api) DatabaseSet(c *gin.Context) {
	var data []byte
	data, err := io.ReadAll(c.Request.Body)

	key := strings.TrimPrefix(c.Param("key"), "/")

	if err == nil {
		api.BadgerSync.Lock()

		err = api.Badger.Update(func(txn *badger.Txn) error {
			err = txn.Set([]byte(key), data)
			return err
		})

		api.BadgerSync.Unlock()

		if err != nil {
			c.JSON(http.StatusInternalServerError, contracts.ResponseOperator{
				Explanation:      "failed to store value in the key-value store",
				ErrorExplanation: err.Error(),
				Error:            true,
				Success:          false,
				Data:             nil,
			})
		} else {
			c.JSON(http.StatusOK, contracts.ResponseOperator{
				Explanation:      "value stored in the key value store",
				ErrorExplanation: "",
				Error:            false,
				Success:          true,
				Data: map[string]any{
					key: string(data),
				},
			})
		}
	} else {
		c.JSON(http.StatusInternalServerError, contracts.ResponseOperator{
			Explanation:      "failed to store value in the key-value store",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})
	}
}

func (api *Api) Propose(c *gin.Context) {
	var data []byte
	data, err := io.ReadAll(c.Request.Body)

	fmt.Println("Propose web!")

	if err != nil {
		c.JSON(http.StatusInternalServerError, contracts.ResponseOperator{
			Explanation:      "failed to store value in the key-value store",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})

		return
	}

	if api.Cluster.KVStore == nil {
		c.JSON(http.StatusInternalServerError, contracts.ResponseOperator{
			Explanation:      "key-value store is not started yet",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})

		return
	}

	key := strings.TrimPrefix(c.Param("key"), "/")
	api.Cluster.KVStore.Propose(key, string(data), api.Config.Agent)

	c.JSON(http.StatusOK, contracts.ResponseImplementation{
		Explanation:      "value stored in the key value store",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data: map[string]any{
			key: string(data),
		},
	})
}

// DatabaseGetKeysPrefix godoc
//
//	@Summary		Get keys by prefix in the key-value store
//	@Description	get all keys by prefix in the key-value store
//	@Tags			database
//	@Produce		json
//	@Success		200	{object}	  contracts.ResponseOperator
//	@Failure		400	{object}	  contracts.ResponseOperator
//	@Failure		404	{object}	  contracts.ResponseOperator
//	@Failure		500	{object}	  contracts.ResponseOperator
//	@Router			/database/{key}/{prefix} [get]
func (api *Api) DatabaseGetKeysPrefix(c *gin.Context) {
	var keys []string

	prefix := strings.TrimPrefix(c.Param("prefix"), "/")

	api.BadgerSync.RLock()

	err := api.Badger.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(prefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()

			keys = append(keys, string(k))
		}

		return nil
	})

	api.BadgerSync.RUnlock()

	if err == nil {
		c.JSON(http.StatusOK, contracts.ResponseOperator{
			Explanation:      "keys found",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
			Data: map[string]any{
				"keys": keys,
			},
		})
	} else {
		c.JSON(http.StatusNotFound, contracts.ResponseOperator{
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
//	@Success		200	{object}	  contracts.ResponseOperator
//	@Failure		400	{object}	  contracts.ResponseOperator
//	@Failure		404	{object}	  contracts.ResponseOperator
//	@Failure		500	{object}	  contracts.ResponseOperator
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
		c.JSON(http.StatusOK, contracts.ResponseOperator{
			Explanation:      "succesfully retrieved keys from the key-value store",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
			Data: map[string]any{
				"keys": keys,
			},
		})
	} else {
		c.JSON(http.StatusNotFound, contracts.ResponseOperator{
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
//	@Success		200	{object}	  contracts.ResponseOperator
//	@Failure		400	{object}	  contracts.ResponseOperator
//	@Failure		404	{object}	  contracts.ResponseOperator
//	@Failure		500	{object}	  contracts.ResponseOperator
//	@Router			/database/keys [delete]
func (api *Api) DatabaseRemoveKeys(c *gin.Context) {
	var keys []string

	prefix := strings.TrimPrefix(c.Param("prefix"), "/")

	api.BadgerSync.Lock()

	err := api.Badger.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false

		it := txn.NewIterator(opts)
		defer it.Close()

		var err error

		prefix := []byte(prefix)
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
		c.JSON(http.StatusOK, contracts.ResponseOperator{
			Explanation:      "succesfully removed keys from the key-value store",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
			Data: map[string]any{
				"keys": keys,
			},
		})
	} else {
		c.JSON(http.StatusNotFound, contracts.ResponseOperator{
			Explanation:      "failed to remove keys from the key-value store",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data: map[string]any{
				"keys": keys,
			},
		})
	}
}
