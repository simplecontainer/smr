package status

import (
	"errors"
	"github.com/hmdsefi/gograph"
	"go.uber.org/zap"
	"time"
)

func NewStatus() *Status {
	return &Status{
		Logger: nil,
	}
}

func (status *Status) CreateGraph() {
	status.StateMachine = gograph.New[*StatusState](gograph.Directed())

	created := gograph.NewVertex(&StatusState{STATUS_CREATED, CATEGORY_PRERUN})
	syncing := gograph.NewVertex(&StatusState{STATUS_SYNCING, CATEGORY_WHILERUN})
	insync := gograph.NewVertex(&StatusState{STATUS_INSYNC, CATEGORY_END})
	backoff := gograph.NewVertex(&StatusState{STATUS_BACKOFF, CATEGORY_END})
	invalidgit := gograph.NewVertex(&StatusState{STATUS_INVALID_GIT, CATEGORY_END})
	invaliddefinitions := gograph.NewVertex(&StatusState{STATUS_INVALID_DEFINITIONS, CATEGORY_END})
	inspecting := gograph.NewVertex(&StatusState{STATUS_INSPECTING, CATEGORY_WHILERUN})
	drifted := gograph.NewVertex(&StatusState{STATUS_DRIFTED, CATEGORY_WHILERUN})
	cloning := gograph.NewVertex(&StatusState{STATUS_CLONING_GIT, CATEGORY_WHILERUN})
	cloned := gograph.NewVertex(&StatusState{STATUS_CLONED_GIT, CATEGORY_WHILERUN})
	pendingDelete := gograph.NewVertex(&StatusState{STATUS_PENDING_DELETE, CATEGORY_END})

	status.StateMachine.AddEdge(created, syncing)
	status.StateMachine.AddEdge(created, invalidgit)

	status.StateMachine.AddEdge(cloning, invalidgit)
	status.StateMachine.AddEdge(cloning, syncing)
	status.StateMachine.AddEdge(cloning, cloned)
	status.StateMachine.AddEdge(cloning, inspecting)

	status.StateMachine.AddEdge(cloned, syncing)
	status.StateMachine.AddEdge(cloned, inspecting)

	status.StateMachine.AddEdge(inspecting, invalidgit)
	status.StateMachine.AddEdge(inspecting, drifted)
	status.StateMachine.AddEdge(inspecting, insync)
	status.StateMachine.AddEdge(inspecting, cloning)

	status.StateMachine.AddEdge(syncing, insync)
	status.StateMachine.AddEdge(syncing, backoff)
	status.StateMachine.AddEdge(syncing, invaliddefinitions)

	status.StateMachine.AddEdge(invalidgit, created)

	status.StateMachine.AddEdge(drifted, syncing)

	status.StateMachine.AddEdge(invaliddefinitions, cloning)
	status.StateMachine.AddEdge(invalidgit, cloning)
	status.StateMachine.AddEdge(insync, cloning)
	status.StateMachine.AddEdge(created, cloning)

	status.StateMachine.AddEdge(drifted, pendingDelete)
	status.StateMachine.AddEdge(insync, pendingDelete)
	status.StateMachine.AddEdge(inspecting, pendingDelete)
	status.StateMachine.AddEdge(syncing, pendingDelete)
	status.StateMachine.AddEdge(created, pendingDelete)
	status.StateMachine.AddEdge(cloning, pendingDelete)
	status.StateMachine.AddEdge(cloned, pendingDelete)
	status.StateMachine.AddEdge(invaliddefinitions, pendingDelete)
	status.StateMachine.AddEdge(invalidgit, pendingDelete)
}

func (status *Status) SetState(state string) error {
	st, err := status.TypeFromString(state)

	if err != nil {
		return err
	}

	status.State = st

	return errors.New("failed to set state")
}

func (status *Status) TransitionState(gitops string, destination string) bool {
	currentVertex := status.StateMachine.GetAllVerticesByID(status.State)

	if len(currentVertex) > 0 {
		edges := status.StateMachine.EdgesOf(currentVertex[0])

		for _, edge := range edges {
			if edge.Destination().Label().State == destination {
				status.Logger.Info("gitops transitioned state",
					zap.String("old-state", status.State.State),
					zap.String("new-state", destination),
					zap.String("gitops", gitops),
				)

				status.PreviousState = status.State
				status.State = edge.Destination().Label()
				status.Reconciling = false
				status.LastUpdate = time.Now()

				return true
			}
		}

		if status.State.State != destination {
			status.Logger.Info("gitops failed to transition state",
				zap.String("old-state", status.State.State),
				zap.String("new-state", destination),
				zap.String("gitops", gitops),
			)

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
