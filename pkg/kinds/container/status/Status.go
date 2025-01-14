package status

import (
	"errors"
	"github.com/hmdsefi/gograph"
	"github.com/simplecontainer/smr/pkg/kinds/container/distributed"
	"github.com/simplecontainer/smr/pkg/logger"
	"go.uber.org/zap"
	"time"
)

func New(ChangeC chan distributed.Container) *Status {
	s := &Status{
		State:      &StatusState{},
		LastUpdate: time.Now(),
		ChangeC:    ChangeC,
	}

	s.CreateGraph()
	return s
}

func (status *Status) CreateGraph() {
	status.StateMachine = gograph.New[*StatusState](gograph.Directed())

	transfering := gograph.NewVertex(&StatusState{STATUS_TRANSFERING, CATEGORY_PRERUN})
	created := gograph.NewVertex(&StatusState{STATUS_CREATED, CATEGORY_PRERUN})
	recreated := gograph.NewVertex(&StatusState{STATUS_RECREATED, CATEGORY_PRERUN})
	prepare := gograph.NewVertex(&StatusState{STATUS_PREPARE, CATEGORY_PRERUN})
	dependsChecking := gograph.NewVertex(&StatusState{STATUS_DEPENDS_CHECKING, CATEGORY_PRERUN})
	dependsSolved := gograph.NewVertex(&StatusState{STATUS_DEPENDS_SOLVED, CATEGORY_PRERUN})
	start := gograph.NewVertex(&StatusState{STATUS_START, CATEGORY_PRERUN})
	readinessChecking := gograph.NewVertex(&StatusState{STATUS_READINESS_CHECKING, CATEGORY_WHILERUN})
	readinessReady := gograph.NewVertex(&StatusState{STATUS_READY, CATEGORY_WHILERUN})
	running := gograph.NewVertex(&StatusState{STATUS_RUNNING, CATEGORY_WHILERUN})
	dead := gograph.NewVertex(&StatusState{STATUS_DEAD, CATEGORY_POSTRUN})
	backoff := gograph.NewVertex(&StatusState{STATUS_BACKOFF, CATEGORY_END})

	dependsFailed := gograph.NewVertex(&StatusState{STATUS_DEPENDS_FAILED, CATEGORY_PRERUN})
	readinessFailed := gograph.NewVertex(&StatusState{STATUS_READINESS_FAILED, CATEGORY_WHILERUN})
	pending := gograph.NewVertex(&StatusState{STATUS_PENDING, CATEGORY_PRERUN})
	runtimePending := gograph.NewVertex(&StatusState{STATUS_RUNTIME_PENDING, CATEGORY_WHILERUN})

	kill := gograph.NewVertex(&StatusState{STATUS_KILL, CATEGORY_WHILERUN})
	forceKill := gograph.NewVertex(&StatusState{STATUS_KILL, CATEGORY_WHILERUN})
	pendingDelete := gograph.NewVertex(&StatusState{STATUS_PENDING_DELETE, CATEGORY_END})

	status.StateMachine.AddEdge(transfering, created)

	status.StateMachine.AddEdge(created, prepare)
	status.StateMachine.AddEdge(created, kill)
	status.StateMachine.AddEdge(created, dead)
	status.StateMachine.AddEdge(created, pendingDelete)

	status.StateMachine.AddEdge(recreated, prepare)
	status.StateMachine.AddEdge(recreated, running)
	status.StateMachine.AddEdge(recreated, kill)
	status.StateMachine.AddEdge(recreated, dead)
	status.StateMachine.AddEdge(recreated, pendingDelete)

	status.StateMachine.AddEdge(prepare, dependsChecking)
	status.StateMachine.AddEdge(prepare, pending)
	status.StateMachine.AddEdge(prepare, pendingDelete)

	status.StateMachine.AddEdge(dependsChecking, dependsSolved)
	status.StateMachine.AddEdge(dependsChecking, dependsFailed)
	status.StateMachine.AddEdge(dependsChecking, pendingDelete)

	status.StateMachine.AddEdge(dependsSolved, start)
	status.StateMachine.AddEdge(dependsChecking, dependsFailed)
	status.StateMachine.AddEdge(dependsSolved, pendingDelete)

	status.StateMachine.AddEdge(start, readinessChecking)
	status.StateMachine.AddEdge(start, dead)
	status.StateMachine.AddEdge(start, kill)
	status.StateMachine.AddEdge(start, pendingDelete)
	status.StateMachine.AddEdge(start, runtimePending)
	status.StateMachine.AddEdge(start, created)

	status.StateMachine.AddEdge(readinessChecking, readinessReady)
	status.StateMachine.AddEdge(readinessChecking, readinessFailed)
	status.StateMachine.AddEdge(readinessChecking, dead)
	status.StateMachine.AddEdge(readinessChecking, pendingDelete)

	status.StateMachine.AddEdge(readinessReady, running)
	status.StateMachine.AddEdge(readinessReady, dead)
	status.StateMachine.AddEdge(readinessReady, pendingDelete)

	status.StateMachine.AddEdge(running, dead)
	status.StateMachine.AddEdge(running, readinessChecking)
	status.StateMachine.AddEdge(running, dead)
	status.StateMachine.AddEdge(running, prepare)
	status.StateMachine.AddEdge(running, backoff)
	status.StateMachine.AddEdge(running, kill)
	status.StateMachine.AddEdge(running, pendingDelete)
	status.StateMachine.AddEdge(running, created)

	status.StateMachine.AddEdge(dead, prepare)
	status.StateMachine.AddEdge(dead, backoff)
	status.StateMachine.AddEdge(dead, pendingDelete)
	status.StateMachine.AddEdge(dead, created)

	status.StateMachine.AddEdge(pending, pendingDelete)
	status.StateMachine.AddEdge(pending, created)
	status.StateMachine.AddEdge(pending, prepare)
	status.StateMachine.AddEdge(pending, dependsChecking)

	status.StateMachine.AddEdge(runtimePending, pendingDelete)
	status.StateMachine.AddEdge(runtimePending, created)
	status.StateMachine.AddEdge(runtimePending, dead)
	status.StateMachine.AddEdge(runtimePending, kill)
	status.StateMachine.AddEdge(runtimePending, prepare)
	status.StateMachine.AddEdge(runtimePending, dependsChecking)

	status.StateMachine.AddEdge(dependsFailed, prepare)
	status.StateMachine.AddEdge(dependsFailed, backoff)
	status.StateMachine.AddEdge(dependsFailed, pendingDelete)
	status.StateMachine.AddEdge(dependsFailed, dead)
	status.StateMachine.AddEdge(dependsFailed, created)

	status.StateMachine.AddEdge(readinessFailed, kill)
	status.StateMachine.AddEdge(readinessFailed, backoff)
	status.StateMachine.AddEdge(readinessFailed, pendingDelete)
	status.StateMachine.AddEdge(readinessFailed, created)

	status.StateMachine.AddEdge(kill, dead)
	status.StateMachine.AddEdge(kill, forceKill)
	status.StateMachine.AddEdge(kill, pendingDelete)

	status.StateMachine.AddEdge(forceKill, dead)
	status.StateMachine.AddEdge(forceKill, pendingDelete)

	status.StateMachine.AddEdge(backoff, pendingDelete)
	status.StateMachine.AddEdge(backoff, created)
}

func (status *Status) SetState(state string) error {
	st, err := status.TypeFromString(state)

	if err != nil {
		return err
	}

	status.State = st

	return errors.New("failed to set state")
}

func (status *Status) TransitionState(group string, container string, destination string) bool {
	currentVertex := status.StateMachine.GetAllVerticesByID(status.State)

	if len(currentVertex) > 0 {
		edges := status.StateMachine.EdgesOf(currentVertex[0])

		for _, edge := range edges {
			if edge.Destination().Label().State == destination {
				logger.Log.Info("container transitioned state",
					zap.String("old-state", status.State.State),
					zap.String("new-state", destination),
					zap.String("group", group),
					zap.String("container", container),
				)

				status.State = edge.Destination().Label()
				status.Reconciling = false
				status.LastUpdate = time.Now()

				status.ChangeC <- distributed.Container{
					Group: group,
					Name:  container,
				}

				return true
			}
		}

		if status.State.State != destination {
			logger.Log.Info("container failed to transition state",
				zap.String("old-state", status.State.State),
				zap.String("new-state", destination),
				zap.String("container", container),
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
