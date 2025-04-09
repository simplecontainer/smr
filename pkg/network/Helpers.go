package network

import (
	"encoding/json"
	jsoniter "github.com/json-iterator/go"
	"github.com/simplecontainer/smr/pkg/logger"
)

func ToJSON(data interface{}) json.RawMessage {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	var marshaled []byte

	switch v := data.(type) {
	case string:
		marshaled = []byte(v)
		break
	default:
		var err error
		marshaled, err = json.Marshal(v)

		if err != nil {
			logger.Log.Error(err.Error())
		}
	}

	return marshaled
}
