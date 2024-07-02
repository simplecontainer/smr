package objects

import (
	"encoding/json"
	"github.com/dgraph-io/badger/v4"
	"github.com/r3labs/diff/v3"
	"github.com/simplecontainer/smr/pkg/database"
	"reflect"
	"strings"
	"time"
)

func New() *Object {
	return &Object{
		Changelog:      diff.Changelog{},
		definition:     map[string]any{},
		definitionByte: make([]byte, 0),
		exists:         false,
		changed:        false,
		created:        time.Now(),
		updated:        time.Now(),
	}
}

func ConvertToMap(jsonData []byte) (map[string]any, error) {
	data := map[string]any{}
	err := json.Unmarshal(jsonData, &data)

	if err != nil {
		return nil, err
	}

	return data, nil
}

func (obj *Object) GetDefinition() map[string]any {
	return obj.definition
}
func (obj *Object) GetDefinitionByte() []byte {
	return obj.definitionByte
}

func (obj *Object) Add(db *badger.DB, format database.FormatStructure, data string) error {
	err := database.Put(db, format.ToString(), string(data))

	if err != nil {
		return err
	}

	format.Key = "updated"
	err = database.Put(db, format.ToString(), time.Now().Format(time.RFC3339))

	timeNow := time.Now()

	if err == nil {
		obj.updated = timeNow
	} else {
		return err
	}

	format.Key = "created"
	err = database.Put(db, format.ToString(), time.Now().Format(time.RFC3339))

	if err == nil {
		obj.created = timeNow
	} else {
		return err
	}

	return err
}

func (obj *Object) Update(db *badger.DB, format database.FormatStructure, data string) error {
	err := database.Put(db, format.ToString(), string(data))

	if err != nil {
		return err
	}

	format.Key = "updated"
	err = database.Put(db, format.ToString(), time.Now().Format(time.RFC3339))

	if err == nil {
		obj.updated = time.Now()
	}

	return err
}

func (obj *Object) Find(db *badger.DB, format database.FormatStructure) error {
	val, err := database.Get(db, format.ToString())

	if err == nil {
		data := make(map[string]any)
		err = json.Unmarshal([]byte(val), &data)

		if err != nil {
			return err
		}

		obj.definition = data
		obj.definitionByte = []byte(val)
	} else {
		return err
	}

	format.Key = "created"

	val, err = database.Get(db, format.ToString())

	if err == nil {
		obj.created, err = time.Parse(time.RFC3339, val)

		if err != nil {
			return err
		}
	} else {
		return err
	}

	format.Key = "updated"

	val, err = database.Get(db, format.ToString())

	if err == nil {
		obj.created, err = time.Parse(time.RFC3339, val)

		if err != nil {
			return err
		}
	} else {
		return err
	}

	obj.changed = false
	obj.exists = true

	return nil
}

func FindMany(db *badger.DB, format database.FormatStructure) (map[string]*Object, error) {
	var objects = make(map[string]*Object)
	objectStrings, err := database.GetPrefix(db, format.ToString())

	if err != nil {
		return nil, err
	}

	var data = make(map[string]any)

	for key, value := range objectStrings {
		if strings.Contains(key, "object") {
			data = map[string]any{}
			err = json.Unmarshal([]byte(value), &data)

			if err != nil {
				return nil, err
			}

			obj := New()
			obj.definition = data
			obj.definitionByte = []byte(value)

			objects[key] = obj
		}
	}

	return objects, nil
}

func (obj *Object) Remove(db *badger.DB, format database.FormatStructure) (bool, error) {
	return database.Delete(db, format.ToBytes())
}

func (obj *Object) Diff(definition string) bool {
	data := make(map[string]any)
	err := json.Unmarshal([]byte(definition), &data)

	if err != nil {
		return true
	}

	var changelog diff.Changelog

	if reflect.DeepEqual(obj.definition, data) {
		obj.changed = false
	} else {
		changelog, _ = diff.Diff(obj.definition, data)
		obj.Changelog = changelog
		obj.changed = true
	}

	return obj.changed
}

func (obj *Object) Exists() bool {
	return obj.exists
}

func (obj *Object) ChangeDetected() bool {
	return obj.changed
}
