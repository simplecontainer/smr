package network

import (
	"encoding/json"
	"github.com/simplecontainer/smr/pkg/logger"
)

func ToJson(data interface{}) json.RawMessage {
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
