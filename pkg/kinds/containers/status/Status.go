package status

import (
	"errors"
	"github.com/hmdsefi/gograph"
	"strings"
	"time"
)

func New() *Status {
	s := &Status{
		State:      &State{},
		Pending:    NewPending(),
		LastUpdate: time.Now(),
	}

	s.CreateGraph()
	return s
}

func (status *Status) CreateGraph() {
	status.StateMachine = gograph.New[*State](gograph.Directed())

	clean := gograph.NewVertex(&State{CLEAN, INITIAL, CATEGORY_CLEAN})
	dependsFailed := gograph.NewVertex(&State{DEPENDS_FAILED, INITIAL, CATEGORY_PRERUN})
	transfering := gograph.NewVertex(&State{TRANSFERING, INITIAL, CATEGORY_PRERUN})
	created := gograph.NewVertex(&State{CREATED, INITIAL, CATEGORY_PRERUN})
	prepare := gograph.NewVertex(&State{PREPARE, INITIAL, CATEGORY_PRERUN})
	init := gograph.NewVertex(&State{INIT, INITIAL, CATEGORY_PRERUN})
	dependsChecking := gograph.NewVertex(&State{DEPENDS_CHECKING, INITIAL, CATEGORY_PRERUN})
	dependsSolved := gograph.NewVertex(&State{DEPENDS_SOLVED, INITIAL, CATEGORY_PRERUN})
	start := gograph.NewVertex(&State{START, INITIAL, CATEGORY_PRERUN})
	restart := gograph.NewVertex(&State{RESTART, INITIAL, CATEGORY_WHILERUN})

	readinessFailed := gograph.NewVertex(&State{READINESS_FAILED, INITIAL, CATEGORY_WHILERUN})
	runtimePending := gograph.NewVertex(&State{RUNTIME_PENDING, INITIAL, CATEGORY_WHILERUN})
	kill := gograph.NewVertex(&State{KILL, INITIAL, CATEGORY_WHILERUN})
	forceKill := gograph.NewVertex(&State{KILL, INITIAL, CATEGORY_WHILERUN})
	readinessChecking := gograph.NewVertex(&State{READINESS_CHECKING, INITIAL, CATEGORY_WHILERUN})
	change := gograph.NewVertex(&State{CHANGE, INITIAL, CATEGORY_WHILERUN})
	readinessReady := gograph.NewVertex(&State{READY, INITIAL, CATEGORY_WHILERUN})

	dead := gograph.NewVertex(&State{DEAD, INITIAL, CATEGORY_POSTRUN})

	running := gograph.NewVertex(&State{RUNNING, INITIAL, CATEGORY_END})
	backoff := gograph.NewVertex(&State{BACKOFF, INITIAL, CATEGORY_END})
	pendingDelete := gograph.NewVertex(&State{PENDING_DELETE, INITIAL, CATEGORY_END})
	daemonFailure := gograph.NewVertex(&State{DAEMON_FAILURE, INITIAL, CATEGORY_END})
	initFailed := gograph.NewVertex(&State{INIT_FAILED, INITIAL, CATEGORY_END})
	pending := gograph.NewVertex(&State{PENDING, INITIAL, CATEGORY_END})

	status.StateMachine.AddEdge(transfering, created)

	status.StateMachine.AddEdge(change, created)

	status.StateMachine.AddEdge(created, change)
	status.StateMachine.AddEdge(created, clean)
	status.StateMachine.AddEdge(created, prepare)
	status.StateMachine.AddEdge(created, kill)
	status.StateMachine.AddEdge(created, dead)
	status.StateMachine.AddEdge(created, pendingDelete)
	status.StateMachine.AddEdge(created, daemonFailure)
	status.StateMachine.AddEdge(created, restart)
	status.StateMachine.AddEdge(created, clean)

	status.StateMachine.AddEdge(clean, prepare)
	status.StateMachine.AddEdge(clean, pendingDelete)
	status.StateMachine.AddEdge(clean, daemonFailure)

	status.StateMachine.AddEdge(restart, clean)

	status.StateMachine.AddEdge(prepare, change)
	status.StateMachine.AddEdge(prepare, dependsChecking)
	status.StateMachine.AddEdge(prepare, pending)
	status.StateMachine.AddEdge(prepare, pendingDelete)
	status.StateMachine.AddEdge(prepare, restart)

	status.StateMachine.AddEdge(dependsChecking, change)
	status.StateMachine.AddEdge(dependsChecking, dependsSolved)
	status.StateMachine.AddEdge(dependsChecking, dependsFailed)
	status.StateMachine.AddEdge(dependsChecking, pendingDelete)
	status.StateMachine.AddEdge(dependsChecking, restart)

	status.StateMachine.AddEdge(dependsSolved, change)
	status.StateMachine.AddEdge(dependsSolved, start)
	status.StateMachine.AddEdge(dependsSolved, init)
	status.StateMachine.AddEdge(dependsChecking, dependsFailed)
	status.StateMachine.AddEdge(dependsSolved, pendingDelete)
	status.StateMachine.AddEdge(dependsSolved, restart)

	status.StateMachine.AddEdge(init, init)
	status.StateMachine.AddEdge(init, initFailed)
	status.StateMachine.AddEdge(init, start)
	status.StateMachine.AddEdge(init, pendingDelete)
	status.StateMachine.AddEdge(init, restart)

	status.StateMachine.AddEdge(start, change)
	status.StateMachine.AddEdge(start, readinessChecking)
	status.StateMachine.AddEdge(start, dead)
	status.StateMachine.AddEdge(start, kill)
	status.StateMachine.AddEdge(start, pendingDelete)
	status.StateMachine.AddEdge(start, runtimePending)
	status.StateMachine.AddEdge(start, created)
	status.StateMachine.AddEdge(start, daemonFailure)
	status.StateMachine.AddEdge(start, restart)

	status.StateMachine.AddEdge(readinessChecking, change)
	status.StateMachine.AddEdge(readinessChecking, readinessReady)
	status.StateMachine.AddEdge(readinessChecking, readinessFailed)
	status.StateMachine.AddEdge(readinessChecking, dead)
	status.StateMachine.AddEdge(readinessChecking, pendingDelete)
	status.StateMachine.AddEdge(readinessChecking, restart)

	status.StateMachine.AddEdge(readinessReady, change)
	status.StateMachine.AddEdge(readinessReady, running)
	status.StateMachine.AddEdge(readinessReady, dead)
	status.StateMachine.AddEdge(readinessReady, pendingDelete)
	status.StateMachine.AddEdge(readinessReady, restart)

	status.StateMachine.AddEdge(running, change)
	status.StateMachine.AddEdge(running, dead)
	status.StateMachine.AddEdge(running, readinessChecking)
	status.StateMachine.AddEdge(running, dead)
	status.StateMachine.AddEdge(running, prepare)
	status.StateMachine.AddEdge(running, backoff)
	status.StateMachine.AddEdge(running, kill)
	status.StateMachine.AddEdge(running, pendingDelete)
	status.StateMachine.AddEdge(running, created)
	status.StateMachine.AddEdge(running, restart)

	status.StateMachine.AddEdge(dead, change)
	status.StateMachine.AddEdge(dead, prepare)
	status.StateMachine.AddEdge(dead, backoff)
	status.StateMachine.AddEdge(dead, pendingDelete)
	status.StateMachine.AddEdge(dead, created)
	status.StateMachine.AddEdge(dead, restart)

	status.StateMachine.AddEdge(pending, change)
	status.StateMachine.AddEdge(pending, pendingDelete)
	status.StateMachine.AddEdge(pending, created)
	status.StateMachine.AddEdge(pending, prepare)
	status.StateMachine.AddEdge(pending, dependsChecking)
	status.StateMachine.AddEdge(pending, restart)

	status.StateMachine.AddEdge(runtimePending, change)
	status.StateMachine.AddEdge(runtimePending, pendingDelete)
	status.StateMachine.AddEdge(runtimePending, created)
	status.StateMachine.AddEdge(runtimePending, dead)
	status.StateMachine.AddEdge(runtimePending, kill)
	status.StateMachine.AddEdge(runtimePending, prepare)
	status.StateMachine.AddEdge(runtimePending, dependsChecking)
	status.StateMachine.AddEdge(runtimePending, restart)

	status.StateMachine.AddEdge(dependsFailed, change)
	status.StateMachine.AddEdge(dependsFailed, prepare)
	status.StateMachine.AddEdge(dependsFailed, backoff)
	status.StateMachine.AddEdge(dependsFailed, pendingDelete)
	status.StateMachine.AddEdge(dependsFailed, dead)
	status.StateMachine.AddEdge(dependsFailed, created)
	status.StateMachine.AddEdge(dependsFailed, restart)

	status.StateMachine.AddEdge(readinessFailed, change)
	status.StateMachine.AddEdge(readinessFailed, kill)
	status.StateMachine.AddEdge(readinessFailed, backoff)
	status.StateMachine.AddEdge(readinessFailed, pendingDelete)
	status.StateMachine.AddEdge(readinessFailed, created)
	status.StateMachine.AddEdge(readinessFailed, restart)

	status.StateMachine.AddEdge(kill, dead)
	status.StateMachine.AddEdge(kill, forceKill)
	status.StateMachine.AddEdge(kill, pendingDelete)
	status.StateMachine.AddEdge(kill, restart)

	status.StateMachine.AddEdge(forceKill, dead)
	status.StateMachine.AddEdge(forceKill, pendingDelete)
	status.StateMachine.AddEdge(forceKill, restart)

	status.StateMachine.AddEdge(backoff, pendingDelete)
	status.StateMachine.AddEdge(backoff, created)
	status.StateMachine.AddEdge(backoff, restart)

	status.StateMachine.AddEdge(daemonFailure, pendingDelete)
	status.StateMachine.AddEdge(daemonFailure, restart)

	status.StateMachine.AddEdge(pendingDelete, clean)
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

	return errors.New("failed to set state")
}

func (status *Status) TransitionState(group string, container string, destination string) bool {
	currentVertex := status.StateMachine.GetAllVerticesByID(status.State)

	if len(currentVertex) > 0 {
		edges := status.StateMachine.EdgesOf(currentVertex[0])

		for _, edge := range edges {
			if edge.Destination().Label().State == destination {
				oldState := strings.Clone(status.State.State)

				status.State = edge.Destination().Label()
				status.State.PreviousState = oldState
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

func (status *Status) TypeFromString(state string) (*State, error) {
	vertexes := status.StateMachine.GetAllVertices()

	for _, v := range vertexes {
		if v.Label().State == state {
			return v.Label(), nil
		}
	}

	return &State{}, errors.New("state not found")
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
