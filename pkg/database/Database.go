package database

import (
	"github.com/dgraph-io/badger/v4"
)

func Put(Badger *badger.DB, key string, value string) error {
	err := Badger.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), []byte(value))

		return err
	})

	return err
}

func Get(Badger *badger.DB, key string) (string, error) {
	var value []byte

	err := Badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		value, err = item.ValueCopy(nil)
		return err
	})

	if err != nil {
		return "", err
	} else {
		return string(value), nil
	}
}

func GetPrefix(Badger *badger.DB, key string) (map[string]string, error) {
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
	err := Badger.DropPrefix(key)

	if err != nil {
		return false, err
	} else {
		return true, nil
	}
}
