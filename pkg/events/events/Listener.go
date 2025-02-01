package events

import (
	"encoding/json"
	"github.com/simplecontainer/smr/pkg/KV"
	"github.com/simplecontainer/smr/pkg/acks"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/logger"
	"go.uber.org/zap"
)

func Listen(kindRegistry map[string]contracts.Kind, e chan KV.KV) {
	for {
		select {
		case data := <-e:
			var event Event

			format, _ := f.NewFromString(data.Key)
			acks.ACKS.Ack(format.GetUUID())

			err := json.Unmarshal(data.Val, &event)

			if err != nil {
				logger.Log.Debug("failed to parse event for processing", zap.String("event", string(data.Val)))
			}

			Handle(kindRegistry, event, data.Node)
		}
	}
}

func Handle(kindRegistry map[string]contracts.Kind, event Event, node uint64) {
	go func() {
		err := kindRegistry[event.Target].Event(event)

		if err != nil {
			logger.Log.Error(err.Error())
		}

		return
	}()
}
