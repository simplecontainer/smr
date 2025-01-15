package distributed

import (
	"encoding/gob"
	"github.com/simplecontainer/smr/pkg/logger"
	"go.uber.org/zap"
)

func NewDecode(decoder *gob.Decoder, node uint64) KV {
	var data KV

	if err := decoder.Decode(&data); err != nil {
		logger.Log.Error("raftexample: could not decode message (%v)", zap.Error(err))
	}

	if node != data.Node {
		data.Local = true
	}

	return data
}

func NewEncode(key string, value []byte, node uint64, category int) KV {
	return KV{
		Key:      key,
		Val:      value,
		Node:     node,
		Category: category,
		Local:    false,
	}
}

func (kv KV) IsLocal() bool {
	return kv.Local
}
