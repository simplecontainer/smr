package objects

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/simplecontainer/smr/pkg/f"
	"net/http"
)

func (obj *Object) FindAndConvert(client *http.Client, format *f.Format, destination interface{}) {
	err := obj.Find(format)
	if err != nil {
		return
	}

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
