package api

import (
	"encoding/json"
	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	"github.com/qdnqn/smr/pkg/database"
	"github.com/qdnqn/smr/pkg/logger"
	"io"
	"net/http"
)

// DatabaseGet godoc
//
//	@Summary		Get value from the key-value store
//	@Description	get string by key from the key-value store
//	@Tags			database
//	@Produce		json
//	@Param			key	path		string	true	"RandomKey"
//	@Success		200	{object}	database.Response
//	@Failure		400	{object}	database.Response
//	@Failure		404	{object}	database.Response
//	@Failure		500	{object}	database.Response
//	@Router			/database/{key} [get]
func (api *Api) DatabaseGet(c *gin.Context) {
	err := api.Badger.View(func(txn *badger.Txn) error {
		var value []byte

		item, err := txn.Get([]byte(c.Param("key")))
		if err != nil {
			c.JSON(http.StatusNotFound, database.Response{
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
			c.JSON(http.StatusNotFound, database.Response{
				Explanation:      "key not found",
				ErrorExplanation: "",
				Error:            true,
				Success:          false,
				Data:             nil,
			})

			return nil
		}

		c.JSON(http.StatusOK, database.Response{
			Explanation:      "found key in the key-value store",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
			Data: map[string]any{
				c.Param("key"): value,
			},
		})

		return nil
	})

	if err != nil {
		logger.Log.Error(err.Error())

		c.JSON(http.StatusNotFound, database.Response{
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
//	@Success		200		{object}	database.Response
//	@Failure		400		{object}	database.Response
//	@Failure		404		{object}	database.Response
//	@Failure		500		{object}	database.Response
//	@Router			/database/{key} [post]
func (api *Api) DatabaseSet(c *gin.Context) {
	jsonData, err := io.ReadAll(c.Request.Body)

	if err == nil {
		valueSent := Kv{}

		if err = json.Unmarshal(jsonData, &valueSent); err != nil {
			c.JSON(http.StatusNotFound, database.Response{
				Explanation:      "failed to store value in the key-value store",
				ErrorExplanation: err.Error(),
				Error:            true,
				Success:          false,
				Data:             nil,
			})

			return
		}

		err = api.Badger.Update(func(txn *badger.Txn) error {
			err = txn.Set([]byte(c.Param("key")), []byte(valueSent.Value))
			return err
		})

		if err != nil {
			c.JSON(http.StatusNotFound, database.Response{
				Explanation:      "failed to store value in the key-value store",
				ErrorExplanation: err.Error(),
				Error:            true,
				Success:          false,
				Data:             nil,
			})
		} else {
			c.JSON(http.StatusOK, database.Response{
				Explanation:      "value stored in the key value store",
				ErrorExplanation: "",
				Error:            false,
				Success:          true,
				Data: map[string]any{
					c.Param(":key"): valueSent.Value,
				},
			})
		}
	} else {
		c.JSON(http.StatusNotFound, database.Response{
			Explanation:      "failed to store value in the key-value store",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})
	}
}

// DatabaseGetKeysPrefix godoc
//
//	@Summary		Get keys by prefix in the key-value store
//	@Description	get all keys by prefix in the key-value store
//	@Tags			database
//	@Produce		json
//	@Success		200	{object}	database.Response
//	@Failure		400	{object}	database.Response
//	@Failure		404	{object}	database.Response
//	@Failure		500	{object}	database.Response
//	@Router			/database/{key}/{prefix} [get]
func (api *Api) DatabaseGetKeysPrefix(c *gin.Context) {
	var keys []string

	err := api.Badger.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(c.Param("prefix"))
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			keys = append(keys, string(k))
		}

		return nil
	})

	if err == nil {
		c.JSON(http.StatusNotFound, database.Response{
			Explanation:      "failed to store value in the key-value store",
			ErrorExplanation: err.Error(),
			Error:            false,
			Success:          true,
			Data: map[string]any{
				"keys": keys,
			},
		})
	} else {
		c.JSON(http.StatusNotFound, database.Response{
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
//	@Success		200	{object}	database.Response
//	@Failure		400	{object}	database.Response
//	@Failure		404	{object}	database.Response
//	@Failure		500	{object}	database.Response
//	@Router			/database/keys [get]
func (api *Api) DatabaseGetKeys(c *gin.Context) {
	var keys []string

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

	if err == nil {
		c.JSON(http.StatusOK, database.Response{
			Explanation:      "succesfully retrieved keys from the key-value store",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
			Data: map[string]any{
				"keys": keys,
			},
		})
	} else {
		c.JSON(http.StatusNotFound, database.Response{
			Explanation:      "failed to retrieve keys from the key-value store",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})
	}
}
