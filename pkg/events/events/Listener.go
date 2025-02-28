package events

import (
	"encoding/json"
	"fmt"
	"github.com/simplecontainer/smr/pkg/KV"
	"github.com/simplecontainer/smr/pkg/acks"
	"github.com/simplecontainer/smr/pkg/contracts/ievents"
	"github.com/simplecontainer/smr/pkg/contracts/ikinds"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/wss"
	"go.uber.org/zap"
)

func Listen(kindRegistry map[string]ikinds.Kind, e chan KV.KV, deleteC map[string]chan ievents.Event, wss *wss.WebSockets) {
	for {
		select {
		case data := <-e:
			var event Event

			format := f.NewFromString(data.Key)
			acks.ACKS.Ack(format.GetUUID())

			err := json.Unmarshal(data.Val, &event)

			if err != nil {
				logger.Log.Debug("failed to parse event for processing", zap.String("event", string(data.Val)))
			}

			wss.Lock.RLock()
			for _, ch := range wss.Channels {
				ch <- event
			}
			wss.Lock.RUnlock()

			Handle(kindRegistry, deleteC, event, data.Node)
		}
	}
}

func Handle(kindRegistry map[string]ikinds.Kind, deleteC map[string]chan ievents.Event, event Event, node uint64) {
	go func() {
		kind, ok := kindRegistry[event.Target]

		if ok {
			if event.GetType() == EVENT_DELETED {
				format := f.New(event.GetPrefix(), event.GetKind(), event.GetGroup(), event.GetName())

				ch, ok := deleteC[format.ToString()]

				if ok {
					ch <- event
				}
			}

			err := kind.Event(event)

			if err != nil {
				logger.Log.Error(err.Error(), zap.String("event", fmt.Sprintf("%s", event)))
			}
		}

		return
	}()
}
