package status

import (
	"errors"
	"github.com/hmdsefi/gograph"
	"time"
)

func New() *Status {
	s := &Status{
		State:      &StatusState{},
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
	pushing_changes := gograph.NewVertex(&StatusState{COMMIT_GIT, CATEGORY_WHILERUN})
	cloning := gograph.NewVertex(&StatusState{CLONING_GIT, CATEGORY_WHILERUN})
	cloned := gograph.NewVertex(&StatusState{CLONED_GIT, CATEGORY_WHILERUN})
	pendingDelete := gograph.NewVertex(&StatusState{DELETE, CATEGORY_END})
	insync := gograph.NewVertex(&StatusState{INSYNC, CATEGORY_END})
	drifted := gograph.NewVertex(&StatusState{DRIFTED, CATEGORY_END})
	backoff := gograph.NewVertex(&StatusState{BACKOFF, CATEGORY_END})
	invalidgit := gograph.NewVertex(&StatusState{INVALID_GIT, CATEGORY_END})
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
	status.StateMachine.AddEdge(drifted, pushing_changes)

	status.StateMachine.AddEdge(invaliddefinitions, cloning)
	status.StateMachine.AddEdge(invalidgit, cloning)
	status.StateMachine.AddEdge(insync, cloning)
	status.StateMachine.AddEdge(created, cloning)

	status.StateMachine.AddEdge(pushing_changes, cloning)
	status.StateMachine.AddEdge(pushing_changes, invalidgit)

	status.StateMachine.AddEdge(insync, inspecting)
	status.StateMachine.AddEdge(insync, pushing_changes)

	status.StateMachine.AddEdge(drifted, pendingDelete)
	status.StateMachine.AddEdge(insync, pendingDelete)
	status.StateMachine.AddEdge(inspecting, pendingDelete)
	status.StateMachine.AddEdge(syncing, pendingDelete)
	status.StateMachine.AddEdge(created, pendingDelete)
	status.StateMachine.AddEdge(cloning, pendingDelete)
	status.StateMachine.AddEdge(cloned, pendingDelete)
	status.StateMachine.AddEdge(invaliddefinitions, pendingDelete)
	status.StateMachine.AddEdge(invalidgit, pendingDelete)
	status.StateMachine.AddEdge(pushing_changes, pendingDelete)

}

func (status *Status) GetPending() *Pending {
	return status.Pending
}

func (status *Status) SetState(state string) error {
	st, err := status.TypeFromString(state)

	if err != nil {
		return err
	}

	status.State = st
	status.LastUpdate = time.Now()

	return errors.New("failed to set state")
}

func (status *Status) TransitionState(group string, name string, destination string) bool {
	currentVertex := status.StateMachine.GetAllVerticesByID(status.State)

	if len(currentVertex) > 0 {
		edges := status.StateMachine.EdgesOf(currentVertex[0])

		for _, edge := range edges {
			if edge.Destination().Label().State == destination {
				status.State = edge.Destination().Label()
				status.Reconciling = false
				status.LastUpdate = time.Now()

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
	return status.State.State
}

func (status *Status) GetCategory() int8 {
	return status.State.category
}

func (status *Status) IfStateIs(state string) bool {
	return status.State.State == state
}
