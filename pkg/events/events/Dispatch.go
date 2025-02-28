package events

import (
	"github.com/simplecontainer/smr/pkg/contracts/ishared"
	"github.com/simplecontainer/smr/pkg/logger"
)

func Dispatch(event Event, shared ishared.Shared, node uint64) {
	if shared.GetManager().Config.KVStore.Node == node {
		err := event.Propose(shared.GetManager().Cluster.KVStore, node)

		if err != nil {
			logger.Log.Error(err.Error())
		}
	}
}

func DispatchGroup(events []Event, shared ishared.Shared, node uint64) {
	if shared.GetManager().Config.KVStore.Node == node {
		for _, event := range events {
			Dispatch(event, shared, node)
		}
	}
}
