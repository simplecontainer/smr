package api

import (
	"encoding/json"
	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/logger"
	"io"
	"net/http"
	"strings"
)

// SecretsGet godoc
//
//	@Summary		Get value from the key-value store
//	@Description	get string by key from the key-value store
//	@Tags			database
//	@Produce		json
//	@Param			key	path		string	true	"RandomKey"
//	@Success		200	{object}	contracts.ResponseOperator
//	@Failure		400	{object}	contracts.ResponseOperator
//	@Failure		404	{object}	contracts.ResponseOperator
//	@Failure		500	{object}	contracts.ResponseOperator
//	@Router			/database/{key} [get]
func (api *Api) SecretsGet(c *gin.Context) {
	if !strings.HasPrefix(c.Param("secret"), "secret.") {
		c.JSON(http.StatusBadRequest, contracts.ResponseOperator{
			Explanation:      "secret not found",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
			Data:             nil,
		})

		return
	}

	err := api.Badger.View(func(txn *badger.Txn) error {
		var value []byte

		item, err := txn.Get([]byte(c.Param("secret")))
		if err != nil {
			c.JSON(http.StatusNotFound, contracts.ResponseOperator{
				Explanation:      "secret not found",
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
				Explanation:      "secret not found",
				ErrorExplanation: "",
				Error:            true,
				Success:          false,
				Data:             nil,
			})

			return nil
		}

		c.JSON(http.StatusOK, contracts.ResponseOperator{
			Explanation:      "found secret in the secret store",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
			Data: map[string]any{
				c.Param("secret"): value,
			},
		})

		return nil
	})

	if err != nil {
		logger.Log.Error(err.Error())

		c.JSON(http.StatusNotFound, contracts.ResponseOperator{
			Explanation:      "failed to read from the secret store",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})
	}
}

// SecretsSet godoc
//
//	@Summary		Set value in the key-value store
//	@Description	set string by key in the key-value store
//	@Tags			database
//	@Accepts		json
//	@Produce		json
//	@Param			key		path		string	true	"RandomKey"
//	@Param			value	body		Kv		true	"value"
//	@Success		200		{object}	contracts.ResponseOperator
//	@Failure		400		{object}	contracts.ResponseOperator
//	@Failure		404		{object}	contracts.ResponseOperator
//	@Failure		500		{object}	contracts.ResponseOperator
//	@Router			/database/{key} [post]
func (api *Api) SecretsSet(c *gin.Context) {
	jsonData, err := io.ReadAll(c.Request.Body)

	if err == nil {
		valueSent := Kv{}

		if err = json.Unmarshal(jsonData, &valueSent); err != nil {
			c.JSON(http.StatusNotFound, contracts.ResponseOperator{
				Explanation:      "failed to store secret in the secret store",
				ErrorExplanation: err.Error(),
				Error:            true,
				Success:          false,
				Data:             nil,
			})

			return
		}

		err = api.Badger.Update(func(txn *badger.Txn) error {
			err = txn.Set([]byte(c.Param("secret")), []byte(valueSent.Value))
			return err
		})

		if err != nil {
			c.JSON(http.StatusNotFound, contracts.ResponseOperator{
				Explanation:      "failed to store secret in the secret store",
				ErrorExplanation: err.Error(),
				Error:            true,
				Success:          false,
				Data:             nil,
			})
		} else {
			c.JSON(http.StatusOK, contracts.ResponseOperator{
				Explanation:      "secret stored in the secret store",
				ErrorExplanation: "",
				Error:            false,
				Success:          true,
				Data: map[string]any{
					c.Param("secret"): valueSent.Value,
				},
			})
		}
	} else {
		c.JSON(http.StatusNotFound, contracts.ResponseOperator{
			Explanation:      "failed to store secret in the secret store",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})
	}
}

// SecretsGetKeys godoc
//
//	@Summary		Get keys by prefix in the key-value store
//	@Description	get all keys by prefix in the key-value store
//	@Tags			database
//	@Produce		json
//	@Success		200	{object}	contracts.ResponseOperator
//	@Failure		400	{object}	contracts.ResponseOperator
//	@Failure		404	{object}	contracts.ResponseOperator
//	@Failure		500	{object}	contracts.ResponseOperator
//	@Router			/database/keys [get]
func (api *Api) SecretsGetKeys(c *gin.Context) {
	var keys []string

	err := api.Badger.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()

			if strings.HasPrefix(string(k), "secret.") {
				keys = append(keys, string(k))
			}
		}

		return nil
	})

	if err == nil {
		c.JSON(http.StatusOK, contracts.ResponseOperator{
			Explanation:      "succesfully retrieved secrets from the secret store",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
			Data: map[string]any{
				"keys": keys,
			},
		})
	} else {
		c.JSON(http.StatusNotFound, contracts.ResponseOperator{
			Explanation:      "failed to retrieve secrets from the secret store",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})
	}
}
