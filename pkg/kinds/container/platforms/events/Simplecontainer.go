package events

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/helpers"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/types"
	"github.com/simplecontainer/smr/pkg/kinds/container/shared"
	"github.com/simplecontainer/smr/pkg/kinds/container/status"
)

func NewEventsListener(shared *shared.Shared, e chan *types.Events) {
	for {
		select {
		case event := <-e:
			Event(shared, event)
		}
	}
}

func Event(shared *shared.Shared, event *types.Events) {
	var container platforms.IContainer

	if event == nil {
		fmt.Println("nil event")
		return
	}

	if helpers.Contains([]string{"Container"}, event.Kind) {
		container = shared.Registry.Find(event.Group, event.Name)
	}

	if container == nil {
		return
	}

	switch event.Kind {
	case "change":
		HandleChange(shared, container)
		break
	default:
		break
	}
}

func HandleChange(shared *shared.Shared, container platforms.IContainer) {
	if !reconcileIgnore(container.GetLabels()) {
		container.GetStatus().TransitionState(container.GetGeneratedName(), status.STATUS_PREPARE)
		shared.Watcher.Find(container.GetGroupIdentifier()).ContainerQueue <- container
	}
}
