package events

import (
	"encoding/json"
	"fmt"
	"github.com/simplecontainer/smr/pkg/KV"
	"github.com/simplecontainer/smr/pkg/acks"
	"github.com/simplecontainer/smr/pkg/contracts/ikinds"
	"github.com/simplecontainer/smr/pkg/distributed"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/wss"
	"go.uber.org/zap"
	"time"
)

func Listen(kindRegistry map[string]ikinds.Kind, e chan KV.KV, informer *distributed.Informer, wss *wss.WebSockets) {
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

			Handle(kindRegistry, informer, event, data.Node)
		}
	}
}

func Handle(kindRegistry map[string]ikinds.Kind, informer *distributed.Informer, event Event, node uint64) {
	kind, ok := kindRegistry[event.Target]

	if ok {
		format := f.New(event.GetPrefix(), event.GetKind(), event.GetGroup(), event.GetName())

		ch := informer.GetCh(format.ToString())

		if ch != nil {
			select {
			case ch <- event:
			case <-time.After(60 * time.Second):
				informer.RmCh(format.ToString())
				logger.Log.Error("informer channel timed out", zap.String("event", fmt.Sprintf("%s", event)))
			}
		}

		err := kind.Event(event)

		if err != nil {
			logger.Log.Error(err.Error(), zap.String("event", fmt.Sprintf("%s", event)))
		}
	}

	return
}
