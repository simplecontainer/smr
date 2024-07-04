package objects

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"net/http"
)

func (obj *Object) FindAndConvert(client *http.Client, format FormatStructure, destination interface{}) {
	obj.Find(client, format)

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
