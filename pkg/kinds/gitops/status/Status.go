package status

import (
	"errors"
	"fmt"
	"github.com/hmdsefi/gograph"
	"time"
)

func New() *Status {
	s := &Status{
		State:      &StatusState{CREATED, CATEGORY_PRERUN},
		StateQueue: make([]*StatusState, 0),
		Pending:    NewPending(),
		LastUpdate: time.Now(),
	}

	s.CreateGraph()
	return s
}

func (status *Status) CreateGraph() {
	status.StateMachine = gograph.New[*StatusState](gograph.Directed())

	created := gograph.NewVertex(&StatusState{CREATED, CATEGORY_PRERUN})
	syncing := gograph.NewVertex(&StatusState{SYNCING, CATEGORY_WHILERUN})
	syncingstate := gograph.NewVertex(&StatusState{SYNCING_STATE, CATEGORY_WHILERUN})
	inspecting := gograph.NewVertex(&StatusState{INSPECTING, CATEGORY_WHILERUN})
	pushingchanges := gograph.NewVertex(&StatusState{COMMIT_GIT, CATEGORY_WHILERUN})
	cloning := gograph.NewVertex(&StatusState{CLONING_GIT, CATEGORY_WHILERUN})
	cloned := gograph.NewVertex(&StatusState{CLONED_GIT, CATEGORY_WHILERUN})
	pushsuccess := gograph.NewVertex(&StatusState{GIT_PUSH_SUCCESS, CATEGORY_WHILERUN})
	pendingdelete := gograph.NewVertex(&StatusState{DELETE, CATEGORY_END})
	insync := gograph.NewVertex(&StatusState{INSYNC, CATEGORY_END})
	drifted := gograph.NewVertex(&StatusState{DRIFTED, CATEGORY_END})
	backoff := gograph.NewVertex(&StatusState{BACKOFF, CATEGORY_END})
	invalidgit := gograph.NewVertex(&StatusState{INVALID_GIT, CATEGORY_END})
	invalidpush := gograph.NewVertex(&StatusState{INVALID_GIT_PUSH, CATEGORY_END})
	invaliddefinitions := gograph.NewVertex(&StatusState{INVALID_DEFINITIONS, CATEGORY_END})

	status.StateMachine.AddEdge(created, syncing)
	status.StateMachine.AddEdge(created, invalidgit)

	status.StateMachine.AddEdge(cloning, invalidgit)
	status.StateMachine.AddEdge(cloning, syncing)
	status.StateMachine.AddEdge(cloning, cloned)
	status.StateMachine.AddEdge(cloning, inspecting)

	status.StateMachine.AddEdge(cloned, syncing)
	status.StateMachine.AddEdge(cloned, inspecting)
	status.StateMachine.AddEdge(cloned, invaliddefinitions)
	status.StateMachine.AddEdge(cloned, cloning)

	status.StateMachine.AddEdge(inspecting, cloned)
	status.StateMachine.AddEdge(inspecting, cloning)
	status.StateMachine.AddEdge(inspecting, drifted)
	status.StateMachine.AddEdge(inspecting, insync)
	status.StateMachine.AddEdge(inspecting, syncingstate)
	status.StateMachine.AddEdge(inspecting, invaliddefinitions)
	status.StateMachine.AddEdge(inspecting, invalidgit)

	status.StateMachine.AddEdge(syncingstate, insync)
	status.StateMachine.AddEdge(syncingstate, drifted)

	status.StateMachine.AddEdge(syncing, insync)
	status.StateMachine.AddEdge(syncing, backoff)
	status.StateMachine.AddEdge(syncing, invaliddefinitions)
	status.StateMachine.AddEdge(syncing, inspecting)

	status.StateMachine.AddEdge(invalidgit, created)

	status.StateMachine.AddEdge(drifted, syncing)
	status.StateMachine.AddEdge(drifted, inspecting)

	status.StateMachine.AddEdge(invaliddefinitions, cloning)
	status.StateMachine.AddEdge(invalidgit, cloning)
	status.StateMachine.AddEdge(insync, cloning)
	status.StateMachine.AddEdge(created, cloning)

	status.StateMachine.AddEdge(pushingchanges, cloning)
	status.StateMachine.AddEdge(pushingchanges, invalidpush)
	status.StateMachine.AddEdge(pushingchanges, pushsuccess)

	status.StateMachine.AddEdge(pushsuccess, cloning)

	status.StateMachine.AddEdge(invalidpush, inspecting)
	status.StateMachine.AddEdge(invalidpush, syncing)

	status.StateMachine.AddEdge(insync, inspecting)
	status.StateMachine.AddEdge(insync, pushingchanges)
	status.StateMachine.AddEdge(insync, cloning)

	status.StateMachine.AddEdge(drifted, pendingdelete)
	status.StateMachine.AddEdge(insync, pendingdelete)
	status.StateMachine.AddEdge(inspecting, pendingdelete)
	status.StateMachine.AddEdge(syncing, pendingdelete)
	status.StateMachine.AddEdge(created, pendingdelete)
	status.StateMachine.AddEdge(cloning, pendingdelete)
	status.StateMachine.AddEdge(cloned, pendingdelete)
	status.StateMachine.AddEdge(invaliddefinitions, pendingdelete)
	status.StateMachine.AddEdge(invalidgit, pendingdelete)
	status.StateMachine.AddEdge(pushingchanges, pendingdelete)
	status.StateMachine.AddEdge(pushsuccess, pendingdelete)
}

func (status *Status) GetPending() *Pending {
	return status.Pending
}

func (status *Status) QueueState(state string) error {
	st, err := status.TypeFromString(state)
	if err != nil {
		return err
	}

	status.mu.Lock()
	defer status.mu.Unlock()

	status.StateQueue = append(status.StateQueue, st)
	return nil
}

// QueueStates adds multiple states to the queue (thread-safe)
func (status *Status) QueueStates(states []string) error {
	status.mu.Lock()
	defer status.mu.Unlock()

	for _, state := range states {
		st, err := status.TypeFromString(state)
		if err != nil {
			return err
		}
		status.StateQueue = append(status.StateQueue, st)
	}
	return nil
}

// PopState removes and returns the first state from the queue (thread-safe)
func (status *Status) PopState() (*StatusState, error) {
	status.mu.Lock()
	defer status.mu.Unlock()

	if len(status.StateQueue) == 0 {
		return nil, errors.New("queue is empty")
	}

	// Get the first state
	state := status.StateQueue[0]

	// Remove it from the queue
	status.StateQueue = status.StateQueue[1:]

	return state, nil
}

// PeekState returns the first state from the queue without removing it (thread-safe)
func (status *Status) PeekState() (*StatusState, error) {
	status.mu.RLock()
	defer status.mu.RUnlock()

	if len(status.StateQueue) == 0 {
		return nil, errors.New("queue is empty")
	}
	return status.StateQueue[0], nil
}

// GetQueueLength returns the current number of states in the queue (thread-safe)
func (status *Status) GetQueueLength() int {
	status.mu.RLock()
	defer status.mu.RUnlock()

	return len(status.StateQueue)
}

func (status *Status) IsQueueEmpty() bool {
	status.mu.RLock()
	defer status.mu.RUnlock()

	return len(status.StateQueue) == 0
}

func (status *Status) ClearQueue() {
	status.mu.Lock()
	defer status.mu.Unlock()

	status.StateQueue = status.StateQueue[:0]
}

func (status *Status) SetState(state string) error {
	st, err := status.TypeFromString(state)
	if err != nil {
		return err
	}

	status.mu.Lock()
	defer status.mu.Unlock()

	status.State = st
	status.LastUpdate = time.Now()
	return nil
}

func (status *Status) TransitionToNext() error {
	status.mu.Lock()
	defer status.mu.Unlock()

	if len(status.StateQueue) == 0 {
		return errors.New("no states in queue to transition to")
	}

	nextState := status.StateQueue[0]

	if !status.canTransitionToUnsafe(nextState.State) {
		return errors.New(fmt.Sprintf("invalid transition from %s to %s", status.State.State, nextState.State))
	}

	status.StateQueue = status.StateQueue[1:]

	status.State = nextState
	status.Reconciling = false
	status.LastUpdate = time.Now()

	return nil
}

func (status *Status) TransitionState(group string, name string, destination string) bool {
	status.mu.Lock()
	defer status.mu.Unlock()

	currentVertex := status.StateMachine.GetAllVerticesByID(status.State)

	if len(currentVertex) > 0 {
		edges := status.StateMachine.EdgesOf(currentVertex[0])

		for _, edge := range edges {
			if edge.Destination().Label().State == destination {
				status.State = edge.Destination().Label()
				status.Reconciling = false
				status.LastUpdate = time.Now()

				// If the destination matches the next state in queue, pop it
				if len(status.StateQueue) > 0 && status.StateQueue[0].State == destination {
					status.StateQueue = status.StateQueue[1:] // Remove the state from queue since we've transitioned to it
				}

				return true
			}
		}

		if status.State.State != destination {
			return false
		}

		return true
	}

	return false
}

func (status *Status) RejectTransition() bool {
	if len(status.StateQueue) > 0 {
		status.StateQueue = status.StateQueue[1:]
		return true
	} else {
		status.StateQueue = []*StatusState{}
		return false
	}
}

func (status *Status) canTransitionTo(destination string) bool {
	status.mu.RLock()
	defer status.mu.RUnlock()

	return status.canTransitionToUnsafe(destination)
}

func (status *Status) canTransitionToUnsafe(destination string) bool {
	if status.State.State == destination {
		return true
	}

	currentVertex := status.StateMachine.GetAllVerticesByID(status.State)

	if len(currentVertex) > 0 {
		edges := status.StateMachine.EdgesOf(currentVertex[0])

		for _, edge := range edges {
			if edge.Destination().Label().State == destination {
				return true
			}
		}
	}

	return false
}

func (status *Status) TypeFromString(state string) (*StatusState, error) {
	vertexes := status.StateMachine.GetAllVertices()

	for _, v := range vertexes {
		if v.Label().State == state {
			return v.Label(), nil
		}
	}

	return &StatusState{}, errors.New("state not found")
}

func (status *Status) GetState() string {
	status.mu.RLock()
	defer status.mu.RUnlock()

	return status.State.State
}

func (status *Status) GetCategory() int8 {
	status.mu.RLock()
	defer status.mu.RUnlock()

	return status.State.category
}

func (status *Status) IfStateIs(state string) bool {
	status.mu.RLock()
	defer status.mu.RUnlock()

	return status.State.State == state
}

func (status *Status) GetStateSnapshot() StatusState {
	status.mu.RLock()
	defer status.mu.RUnlock()

	return StatusState{
		State:    status.State.State,
		category: status.State.category,
	}
}

func (status *Status) GetQueueSnapshot() []*StatusState {
	status.mu.RLock()
	defer status.mu.RUnlock()

	// Create a deep copy of the queue
	snapshot := make([]*StatusState, len(status.StateQueue))
	for i, state := range status.StateQueue {
		snapshot[i] = &StatusState{
			State:    state.State,
			category: state.category,
		}
	}

	return snapshot
}
