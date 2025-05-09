package reconcile

import (
	"errors"
	"github.com/go-playground/assert/v2"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/configuration"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	mock_platforms "github.com/simplecontainer/smr/pkg/kinds/containers/mock"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/readiness"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/state"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/types"
	"github.com/simplecontainer/smr/pkg/kinds/containers/reconcile/mock"
	"github.com/simplecontainer/smr/pkg/kinds/containers/status"
	"github.com/simplecontainer/smr/pkg/kinds/containers/watcher"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/node"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestFromInitialStateToRunning(t *testing.T) {
	type Wanted struct{}

	type Parameters struct{}

	ctrl := gomock.NewController(t)
	registryMock := mock_platforms.NewMockRegistry(ctrl)
	containerMock := mock_platforms.NewMockIContainer(ctrl)

	shared := mock.GetShared(registryMock)

	statusT := status.New()
	statusT.SetState(status.CREATED)

	engineState := state.State{
		Error: "",
		State: "",
	}

	registry := make(map[string]platforms.IContainer)

	testCases := []struct {
		name       string
		mockFunc   func()
		wanted     Wanted
		parameters Parameters
	}{
		{
			"TestFromInitialStateToRunning ",
			func() {
				registryMock.EXPECT().AddOrUpdate("internal", "internal-testing-1", gomock.Any()).DoAndReturn(func(group string, name string, container platforms.IContainer) error {
					registry[name] = container
					return nil
				}).AnyTimes()

				registryMock.EXPECT().Find("", "internal", "internal-testing-1").DoAndReturn(func(prefix string, group string, name string) platforms.IContainer {
					return registry[name]
				}).AnyTimes()

				registryMock.EXPECT().FindGroup("", "internal").DoAndReturn(func(prefix string, group string) []platforms.IContainer {
					return []platforms.IContainer{}
				}).AnyTimes()

				registryMock.EXPECT().Sync("internal", "internal-testing-1").DoAndReturn(func(group string, name string) error {
					return nil
				}).AnyTimes()

				registryMock.EXPECT().Remove("", "internal", "internal-testing-1").DoAndReturn(func(prefix string, group string, name string) error {
					delete(registry, name)
					return nil
				}).AnyTimes()

				containerMock.EXPECT().GetGroup().Return("internal").AnyTimes()
				containerMock.EXPECT().GetGroupIdentifier().Return("internal.internal-testing-1").AnyTimes()
				containerMock.EXPECT().GetGeneratedName().Return("internal-testing-1").AnyTimes()
				containerMock.EXPECT().GetDefinition().Return(mock.DefinitionTestInitial("internal-testing-1", "docker")).AnyTimes()
				containerMock.EXPECT().GetInitDefinition().Return(v1.ContainersInternal{}).AnyTimes()
				containerMock.EXPECT().GetRuntime().DoAndReturn(func() *types.Runtime {
					n := node.NewNode()
					n.NodeID = 1
					n.NodeName = "node-1"

					return &types.Runtime{
						Node: n,
					}
				}).AnyTimes()
				containerMock.EXPECT().GetStatus().DoAndReturn(func() *status.Status {
					return statusT
				}).AnyTimes()

				containerMock.EXPECT().GetReadiness().DoAndReturn(func() []*readiness.Readiness {
					return nil
				}).AnyTimes()

				containerMock.EXPECT().GetState().DoAndReturn(func() (state.State, error) {
					return engineState, nil
				}).AnyTimes()

				containerMock.EXPECT().GetName().DoAndReturn(func() string {
					return "testing"
				}).AnyTimes()

				containerMock.EXPECT().PreRun(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(config *configuration.Configuration, client *clients.Http, user *authentication.User) error {
					return nil
				}).AnyTimes()

				containerMock.EXPECT().Clean().DoAndReturn(func() error {
					return nil
				}).AnyTimes()

				containerMock.EXPECT().Run().DoAndReturn(func() error {
					return nil
				}).AnyTimes()

				containerMock.EXPECT().Stop(gomock.Any()).DoAndReturn(func(signal string) error {
					return nil
				}).AnyTimes()

				containerMock.EXPECT().Delete().DoAndReturn(func() error {
					return nil
				}).AnyTimes()

				containerMock.EXPECT().PostRun(gomock.Any(), gomock.Any()).DoAndReturn(func(config any, dnsCache any) error {
					return nil
				}).AnyTimes()
			},
			Wanted{},
			Parameters{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()

			shared.Registry.AddOrUpdate(containerMock.GetGroup(), containerMock.GetGeneratedName(), containerMock)

			w := watcher.New(containerMock, status.CREATED, shared.User)
			w.Logger = logger.NewLogger("debug", []string{"/dev/stdout"}, []string{"/dev/stdout"})

			// sniffer to implement various scenarios on reconciler
			go func() {
				for {
					select {
					case containerObj := <-w.ContainerQueue:
						if containerObj == nil {
							return
						}

						if containerObj.GetStatus().State.State == status.RUNNING {
							containerObj.GetStatus().TransitionState("", "", status.CREATED)
						}

						if containerObj.GetStatus().State.State == status.CLEAN {
							go func() {
								containerObj.GetStatus().TransitionState("", "", status.DELETE)
							}()

							time.Sleep(1 * time.Second)
						}

						if containerObj.GetStatus().State.State == status.READINESS_CHECKING {
							go func() {
								engineState.State = "running"
								w.ContainerQueue <- containerObj
							}()
						} else {
							w.ContainerQueue <- containerObj
						}
						break
					case containerObj := <-w.DeleteC:
						w.DeleteC <- containerObj
						return
					}
				}
			}()

			go Containers(shared, w)
			HandleTickerAndEvents(shared, w, func(w *watcher.Container) error {
				return errors.New("done")
			})

			assert.Equal(t, statusT.GetState(), status.PENDING_DELETE)
		})
	}
}
