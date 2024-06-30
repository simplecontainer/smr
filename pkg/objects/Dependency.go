package objects

import (
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/mitchellh/mapstructure"
	"github.com/qdnqn/smr/pkg/database"
)

func (obj *Object) FindAndConvert(db *badger.DB, format database.FormatStructure, destination interface{}) {
	obj.Find(db, format)

	if obj.Exists() {
		data := make(map[string]interface{})
		err := json.Unmarshal(obj.GetDefinitionByte(), &data)
		if err != nil {
			panic(err)
		}

		mapstructure.Decode(data, destination)

		fmt.Println(destination)
	} else {
		destination = nil
	}
}
