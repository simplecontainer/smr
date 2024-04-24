package database

import (
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"smr/pkg/logger"
)

func Put(Badger *badger.DB, key string, value string) error {
	logger.Log.Info(fmt.Sprintf("Trying to save into db %s=%s", key, value))
	err := Badger.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), []byte(value))
		return err
	})

	return err
}

func Get(Badger *badger.DB, key string) (string, error) {
	logger.Log.Info(fmt.Sprintf("Trying to get from db %s", key))
	var value []byte

	err := Badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		value, err = item.ValueCopy(nil)
		if err != nil {
			return errors.New(fmt.Sprintf("failed to read %s", err.Error()))
		}

		return nil
	})

	if err != nil {
		return "", err
	} else {
		return string(value), nil
	}
}
