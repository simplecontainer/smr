package reconcile

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/state"
	"github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/kinds/containers/status"
	"github.com/simplecontainer/smr/pkg/kinds/containers/watcher"
	"go.uber.org/zap"
	"sync"
	"time"
)

func Containers(shared *shared.Shared, containerWatcher *watcher.Container, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	containerObj := containerWatcher.Container

	if containerObj.GetStatus().Reconciling {
		containerWatcher.Logger.Info("container already reconciling, waiting for the free slot")
		return
	}

	containerObj.GetStatus().Reconciling = true

	state := GetState(containerWatcher)

	existing := shared.Registry.Find(containerObj.GetDefinition().GetPrefix(), containerObj.GetGroup(), containerObj.GetGeneratedName())
	newState, reconcile := Reconcile(shared, containerWatcher, existing, state.State, state.Error)

	containerObj.GetStatus().Reconciling = false

	fmt.Println(containerObj.GetGeneratedName(), containerObj.GetStatus().State.PreviousState, containerObj.GetStatus().GetState(), newState)

	if newState == "" {
		// Wait till external awakes this reconciler
		return
	}

	// Do not touch container on this node since it is active on another node
	if containerObj.GetStatus().State.State != status.STATUS_TRANSFERING {
		if containerObj.GetStatus().State.State != newState {
			transitioned := containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetName(), newState)

			if !transitioned {
				containerWatcher.Logger.Error("failed to transition state",
					zap.String("old", containerObj.GetStatus().State.State),
					zap.String("new", newState))
			}
		}

		state = GetState(containerWatcher)

		err := shared.Registry.Sync(containerObj.GetGroup(), containerObj.GetGeneratedName())

		if err != nil {
			containerWatcher.Logger.Error(err.Error())
		}

		DispatchEventChange(shared, containerWatcher.Container)

		if reconcile {
			// This is to prevent deadlock since parent of Containers() is waiting for wg.Done()
			go func() {
				if !containerWatcher.Done {
					containerWatcher.ContainerQueue <- containerObj
				}
			}()
		} else {
			if containerObj.GetStatus().GetCategory() == status.CATEGORY_END || containerObj.GetStatus().GetState() == status.STATUS_RUNNING {
				if !containerWatcher.Done {
					containerWatcher.PauseC <- containerObj
				}

				// Skip reconcile chains and inform the gitops after actions are done
				if !containerWatcher.Container.GetStatus().PendingDelete {
					fmt.Println("informing gitops", containerObj.GetGeneratedName(), containerObj.GetStatus().State.PreviousState, containerObj.GetStatus().GetState(), newState)
					DispatchEventInspect(shared, containerWatcher.Container)
				}
			} else {
				containerWatcher.Ticker.Reset(5 * time.Second)
			}
		}
	}
}

func GetState(containerWatcher *watcher.Container) state.State {
	engine, err := containerWatcher.Container.GetState()

	if err != nil {
		return state.State{}
	}

	return engine
}
