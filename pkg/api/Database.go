package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"smr/pkg/logger"
)

func (api *Api) DatabaseGet(c *gin.Context) {
	err := api.Badger.View(func(txn *badger.Txn) error {
		var value []byte

		item, err := txn.Get([]byte(c.Param("key")))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "Key not found",
			})

			return errors.New(fmt.Sprintf("failed to read %s", err.Error()))
		}

		value, err = item.ValueCopy(nil)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "Key not found",
			})

			return errors.New(fmt.Sprintf("failed to read %s", err.Error()))
		}

		c.JSON(http.StatusOK, gin.H{
			c.Param("key"): value,
		})

		return nil
	})

	if err != nil {
		logger.Log.Error("failed reading key from the database")
	}
}

func (api *Api) DatabaseSet(c *gin.Context) {
	jsonData, err := io.ReadAll(c.Request.Body)

	if err == nil {
		valueSent := Kv{}

		if err := json.Unmarshal(jsonData, &valueSent); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Send valid format. eg: {\"value\":\"any value here\"}",
			})
		}

		err := api.Badger.Update(func(txn *badger.Txn) error {
			err := txn.Set([]byte(c.Param("key")), []byte(valueSent.Value))
			return err
		})

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid json sent",
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				c.Param("key"): valueSent.Value,
			})
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid json sent",
		})
	}
}
