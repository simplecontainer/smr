package status

import (
	"fmt"
	"github.com/hmdsefi/gograph"
)

func NewStatus() *Status {
	return &Status{}
}

func (status *Status) CreateGraph() {
	status.StateMachine = gograph.New[string](gograph.Directed())

	created := gograph.NewVertex(STATUS_CREATED)
	running := gograph.NewVertex(STATUS_RUNNING)
	dead := gograph.NewVertex(STATUS_DEAD)
	killed := gograph.NewVertex(STATUS_KILLED)
	backoff := gograph.NewVertex(STATUS_BACKOFF)
	drifted := gograph.NewVertex(STATUS_DRIFTED)
	reconciling := gograph.NewVertex(STATUS_RECONCILING)

	pendingDelete := gograph.NewVertex(STATUS_PENDING_DELETE)

	dependsSolving := gograph.NewVertex(STATUS_DEPENDS_SOLVING)
	dependsSolved := gograph.NewVertex(STATUS_DEPENDS_SOLVED)
	dependsFailed := gograph.NewVertex(STATUS_DEPENDS_FAILED)
	readinesSolving := gograph.NewVertex(STATUS_READINESS)
	readinessReady := gograph.NewVertex(STATUS_READY)
	readinessFailed := gograph.NewVertex(STATUS_READINESS_FAILED)

	status.StateMachine.AddEdge(created, dependsSolving)

	status.StateMachine.AddEdge(drifted, created)

	status.StateMachine.AddEdge(dependsSolving, dependsSolved)
	status.StateMachine.AddEdge(dependsSolving, dependsFailed)

	status.StateMachine.AddEdge(dependsSolved, running)
	status.StateMachine.AddEdge(dependsFailed, dead)

	status.StateMachine.AddEdge(running, dead)
	status.StateMachine.AddEdge(running, readinesSolving)

	status.StateMachine.AddEdge(dead, created)
	status.StateMachine.AddEdge(dead, running)
	status.StateMachine.AddEdge(dead, backoff)
	status.StateMachine.AddEdge(dead, reconciling)

	status.StateMachine.AddEdge(reconciling, running)
	status.StateMachine.AddEdge(reconciling, backoff)

	status.StateMachine.AddEdge(readinesSolving, readinessReady)
	status.StateMachine.AddEdge(readinesSolving, readinessFailed)

	status.StateMachine.AddEdge(readinessReady, running)
	status.StateMachine.AddEdge(created, dead)
	status.StateMachine.AddEdge(drifted, dead)
	status.StateMachine.AddEdge(dependsSolving, dead)
	status.StateMachine.AddEdge(dependsSolved, dead)
	status.StateMachine.AddEdge(dependsFailed, dead)
	status.StateMachine.AddEdge(dead, dead)
	status.StateMachine.AddEdge(readinessFailed, dead)
	status.StateMachine.AddEdge(readinessReady, dead)
	status.StateMachine.AddEdge(readinesSolving, dead)
	status.StateMachine.AddEdge(running, dead)

	status.StateMachine.AddEdge(created, killed)
	status.StateMachine.AddEdge(drifted, killed)
	status.StateMachine.AddEdge(dependsSolving, killed)
	status.StateMachine.AddEdge(dependsSolved, killed)
	status.StateMachine.AddEdge(dependsFailed, killed)
	status.StateMachine.AddEdge(dead, killed)
	status.StateMachine.AddEdge(readinessFailed, killed)
	status.StateMachine.AddEdge(readinessReady, killed)
	status.StateMachine.AddEdge(readinesSolving, killed)
	status.StateMachine.AddEdge(running, killed)

	status.StateMachine.AddEdge(created, pendingDelete)
	status.StateMachine.AddEdge(backoff, pendingDelete)
	status.StateMachine.AddEdge(drifted, pendingDelete)
	status.StateMachine.AddEdge(dependsSolving, pendingDelete)
	status.StateMachine.AddEdge(dependsSolved, pendingDelete)
	status.StateMachine.AddEdge(dependsFailed, pendingDelete)
	status.StateMachine.AddEdge(dead, pendingDelete)
	status.StateMachine.AddEdge(readinessFailed, pendingDelete)
	status.StateMachine.AddEdge(readinessReady, pendingDelete)
	status.StateMachine.AddEdge(readinesSolving, pendingDelete)
	status.StateMachine.AddEdge(running, pendingDelete)
}

func (status *Status) SetState(state string) {
	fmt.Println(fmt.Sprintf("Setting state to the %s", state))
	status.State = state
}

func (status *Status) TransitionState(destination string) bool {
	currentVertex := status.StateMachine.GetAllVerticesByID(status.State)

	if len(currentVertex) > 0 {
		edges := status.StateMachine.EdgesOf(currentVertex[0])

		for _, edge := range edges {
			if edge.Destination().Label() == destination {
				fmt.Println(fmt.Sprintf("State transitioned from %s to %s", status.State, destination))
				status.State = destination
			}
		}

		if status.State != destination {
			fmt.Println(fmt.Sprintf("Failed to transition from %s to %s", status.State, destination))
			return false
		}

		return true
	}

	return false
}

func (status *Status) GetState() string {
	return status.State
}

func (status *Status) IfStateIs(state string) bool {
	return status.State == state
}
