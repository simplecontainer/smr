package smaps

import (
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func (smap *Smap) MarshalJSON() ([]byte, error) {
	tmpMap := make(map[interface{}]interface{})
	smap.Map.Range(func(k, v interface{}) bool {
		tmpMap[k] = v
		return true
	})

	return json.Marshal(tmpMap)
}

func (smap *Smap) UnmarshalJSON(bytes []byte) error {
	var tmpMap map[interface{}]interface{}

	if err := json.Unmarshal(bytes, &tmpMap); err != nil {
		return err
	}

	for key, value := range tmpMap {
		smap.Map.Store(key, value)
	}
	return nil
}
