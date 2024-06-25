package database

import (
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/qdnqn/smr/pkg/logger"
)

func Put(Badger *badger.DB, key string, value string) error {
	logger.Log.Info(fmt.Sprintf("saving into the key-value store %s=%s", key, value))
	err := Badger.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), []byte(value))

		if err == nil {
			logger.Log.Info(fmt.Sprintf("saved into key-value store %s=%s", key, value))
		}

		return err
	})

	return err
}

func Get(Badger *badger.DB, key string) (string, error) {
	logger.Log.Info(fmt.Sprintf("getting from the key-value store %s", key))
	var value []byte

	err := Badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		value, err = item.ValueCopy(nil)
		if err != nil {
			return errors.New(fmt.Sprintf("failed to get %s", err.Error()))
		}

		return nil
	})

	if err != nil {
		return "", err
	} else {
		return string(value), nil
	}
}

func GetPrefix(Badger *badger.DB, key string) (map[string]string, error) {
	logger.Log.Info(fmt.Sprintf("getting from the key-value store %s", key))
	var value = make(map[string]string)

	err := Badger.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek([]byte(key)); it.ValidForPrefix([]byte(key)); it.Next() {
			item := it.Item()
			err := item.Value(func(v []byte) error {
				value[string(item.Key())] = string(v)
				return nil
			})

			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return map[string]string{}, err
	} else {
		return value, nil
	}
}

func Delete(Badger *badger.DB, key []byte) (bool, error) {
	logger.Log.Info(fmt.Sprintf("removing from the key-value store %s", key))

	err := Badger.DropPrefix(key)

	if err != nil {
		return false, err
	} else {
		return true, nil
	}
}
