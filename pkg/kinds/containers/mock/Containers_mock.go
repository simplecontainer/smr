// Code generated by MockGen. DO NOT EDIT.
// Source: platforms/Interface.go
//
// Generated by this command:
//
//	mockgen -source=platforms/Interface.go
//

// Package mock_platforms is a generated GoMock package.
package mock_platforms

import (
	io "io"
	reflect "reflect"

	authentication "github.com/simplecontainer/smr/pkg/authentication"
	client "github.com/simplecontainer/smr/pkg/client"
	configuration "github.com/simplecontainer/smr/pkg/configuration"
	contracts "github.com/simplecontainer/smr/pkg/contracts"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	dns "github.com/simplecontainer/smr/pkg/dns"
	platforms "github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	state "github.com/simplecontainer/smr/pkg/kinds/containers/platforms/state"
	types "github.com/simplecontainer/smr/pkg/kinds/containers/platforms/types"
	status "github.com/simplecontainer/smr/pkg/kinds/containers/status"
	gomock "go.uber.org/mock/gomock"
)

// MockIContainer is a mock of IContainer interface.
type MockIContainer struct {
	ctrl     *gomock.Controller
	recorder *MockIContainerMockRecorder
	isgomock struct{}
}

// MockIContainerMockRecorder is the mock recorder for MockIContainer.
type MockIContainerMockRecorder struct {
	mock *MockIContainer
}

// NewMockIContainer creates a new mock instance.
func NewMockIContainer(ctrl *gomock.Controller) *MockIContainer {
	mock := &MockIContainer{ctrl: ctrl}
	mock.recorder = &MockIContainerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIContainer) EXPECT() *MockIContainerMockRecorder {
	return m.recorder
}

// Delete mocks base method.
func (m *MockIContainer) Delete() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete")
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockIContainerMockRecorder) Delete() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockIContainer)(nil).Delete))
}

// Exec mocks base method.
func (m *MockIContainer) Exec(command []string) (types.ExecResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Exec", command)
	ret0, _ := ret[0].(types.ExecResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exec indicates an expected call of Exec.
func (mr *MockIContainerMockRecorder) Exec(command any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exec", reflect.TypeOf((*MockIContainer)(nil).Exec), command)
}

// GetDefinition mocks base method.
func (m *MockIContainer) GetDefinition() contracts.IDefinition {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDefinition")
	ret0, _ := ret[0].(contracts.IDefinition)
	return ret0
}

// GetDefinition indicates an expected call of GetDefinition.
func (mr *MockIContainerMockRecorder) GetDefinition() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDefinition", reflect.TypeOf((*MockIContainer)(nil).GetDefinition))
}

// GetDomain mocks base method.
func (m *MockIContainer) GetDomain(network string) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDomain", network)
	ret0, _ := ret[0].(string)
	return ret0
}

// GetDomain indicates an expected call of GetDomain.
func (mr *MockIContainerMockRecorder) GetDomain(network any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDomain", reflect.TypeOf((*MockIContainer)(nil).GetDomain), network)
}

// GetGeneratedName mocks base method.
func (m *MockIContainer) GetGeneratedName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGeneratedName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetGeneratedName indicates an expected call of GetGeneratedName.
func (mr *MockIContainerMockRecorder) GetGeneratedName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGeneratedName", reflect.TypeOf((*MockIContainer)(nil).GetGeneratedName))
}

// GetGroup mocks base method.
func (m *MockIContainer) GetGroup() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGroup")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetGroup indicates an expected call of GetGroup.
func (mr *MockIContainerMockRecorder) GetGroup() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGroup", reflect.TypeOf((*MockIContainer)(nil).GetGroup))
}

// GetGroupIdentifier mocks base method.
func (m *MockIContainer) GetGroupIdentifier() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGroupIdentifier")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetGroupIdentifier indicates an expected call of GetGroupIdentifier.
func (mr *MockIContainerMockRecorder) GetGroupIdentifier() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGroupIdentifier", reflect.TypeOf((*MockIContainer)(nil).GetGroupIdentifier))
}

// GetHeadlessDomain mocks base method.
func (m *MockIContainer) GetHeadlessDomain(network string) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetHeadlessDomain", network)
	ret0, _ := ret[0].(string)
	return ret0
}

// GetHeadlessDomain indicates an expected call of GetHeadlessDomain.
func (mr *MockIContainerMockRecorder) GetHeadlessDomain(network any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHeadlessDomain", reflect.TypeOf((*MockIContainer)(nil).GetHeadlessDomain), network)
}

// GetId mocks base method.
func (m *MockIContainer) GetId() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetId")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetId indicates an expected call of GetId.
func (mr *MockIContainerMockRecorder) GetId() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetId", reflect.TypeOf((*MockIContainer)(nil).GetId))
}

// GetInit mocks base method.
func (m *MockIContainer) GetInit() platforms.IPlatform {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInit")
	ret0, _ := ret[0].(platforms.IPlatform)
	return ret0
}

// GetInit indicates an expected call of GetInit.
func (mr *MockIContainerMockRecorder) GetInit() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInit", reflect.TypeOf((*MockIContainer)(nil).GetInit))
}

// GetInitDefinition mocks base method.
func (m *MockIContainer) GetInitDefinition() v1.ContainersInternal {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInitDefinition")
	ret0, _ := ret[0].(v1.ContainersInternal)
	return ret0
}

// GetInitDefinition indicates an expected call of GetInitDefinition.
func (mr *MockIContainerMockRecorder) GetInitDefinition() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInitDefinition", reflect.TypeOf((*MockIContainer)(nil).GetInitDefinition))
}

// GetLabels mocks base method.
func (m *MockIContainer) GetLabels() map[string]string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLabels")
	ret0, _ := ret[0].(map[string]string)
	return ret0
}

// GetLabels indicates an expected call of GetLabels.
func (mr *MockIContainerMockRecorder) GetLabels() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLabels", reflect.TypeOf((*MockIContainer)(nil).GetLabels))
}

// GetName mocks base method.
func (m *MockIContainer) GetName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetName indicates an expected call of GetName.
func (mr *MockIContainerMockRecorder) GetName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetName", reflect.TypeOf((*MockIContainer)(nil).GetName))
}

// GetNode mocks base method.
func (m *MockIContainer) GetNode() uint64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNode")
	ret0, _ := ret[0].(uint64)
	return ret0
}

// GetNode indicates an expected call of GetNode.
func (mr *MockIContainerMockRecorder) GetNode() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNode", reflect.TypeOf((*MockIContainer)(nil).GetNode))
}

// GetNodeName mocks base method.
func (m *MockIContainer) GetNodeName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNodeName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetNodeName indicates an expected call of GetNodeName.
func (mr *MockIContainerMockRecorder) GetNodeName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNodeName", reflect.TypeOf((*MockIContainer)(nil).GetNodeName))
}

// GetRuntime mocks base method.
func (m *MockIContainer) GetRuntime() *types.Runtime {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRuntime")
	ret0, _ := ret[0].(*types.Runtime)
	return ret0
}

// GetRuntime indicates an expected call of GetRuntime.
func (mr *MockIContainerMockRecorder) GetRuntime() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRuntime", reflect.TypeOf((*MockIContainer)(nil).GetRuntime))
}

// GetState mocks base method.
func (m *MockIContainer) GetState() (state.State, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetState")
	ret0, _ := ret[0].(state.State)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetState indicates an expected call of GetState.
func (mr *MockIContainerMockRecorder) GetState() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetState", reflect.TypeOf((*MockIContainer)(nil).GetState))
}

// GetStatus mocks base method.
func (m *MockIContainer) GetStatus() *status.Status {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStatus")
	ret0, _ := ret[0].(*status.Status)
	return ret0
}

// GetStatus indicates an expected call of GetStatus.
func (mr *MockIContainerMockRecorder) GetStatus() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStatus", reflect.TypeOf((*MockIContainer)(nil).GetStatus))
}

// HasDependencyOn mocks base method.
func (m *MockIContainer) HasDependencyOn(arg0, arg1, arg2 string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HasDependencyOn", arg0, arg1, arg2)
	ret0, _ := ret[0].(bool)
	return ret0
}

// HasDependencyOn indicates an expected call of HasDependencyOn.
func (mr *MockIContainerMockRecorder) HasDependencyOn(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasDependencyOn", reflect.TypeOf((*MockIContainer)(nil).HasDependencyOn), arg0, arg1, arg2)
}

// HasOwner mocks base method.
func (m *MockIContainer) HasOwner() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HasOwner")
	ret0, _ := ret[0].(bool)
	return ret0
}

// HasOwner indicates an expected call of HasOwner.
func (mr *MockIContainerMockRecorder) HasOwner() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasOwner", reflect.TypeOf((*MockIContainer)(nil).HasOwner))
}

// InitContainer mocks base method.
func (m *MockIContainer) InitContainer(definitions v1.ContainersInternal, config *configuration.Configuration, client *client.Http, user *authentication.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InitContainer", definitions, config, client, user)
	ret0, _ := ret[0].(error)
	return ret0
}

// InitContainer indicates an expected call of InitContainer.
func (mr *MockIContainerMockRecorder) InitContainer(definitions, config, client, user any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InitContainer", reflect.TypeOf((*MockIContainer)(nil).InitContainer), definitions, config, client, user)
}

// IsGhost mocks base method.
func (m *MockIContainer) IsGhost() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsGhost")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsGhost indicates an expected call of IsGhost.
func (mr *MockIContainerMockRecorder) IsGhost() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsGhost", reflect.TypeOf((*MockIContainer)(nil).IsGhost))
}

// Kill mocks base method.
func (m *MockIContainer) Kill(signal string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Kill", signal)
	ret0, _ := ret[0].(error)
	return ret0
}

// Kill indicates an expected call of Kill.
func (mr *MockIContainerMockRecorder) Kill(signal any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Kill", reflect.TypeOf((*MockIContainer)(nil).Kill), signal)
}

// Logs mocks base method.
func (m *MockIContainer) Logs(arg0 bool) (io.ReadCloser, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Logs", arg0)
	ret0, _ := ret[0].(io.ReadCloser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Logs indicates an expected call of Logs.
func (mr *MockIContainerMockRecorder) Logs(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Logs", reflect.TypeOf((*MockIContainer)(nil).Logs), arg0)
}

// MountResources mocks base method.
func (m *MockIContainer) MountResources() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MountResources")
	ret0, _ := ret[0].(error)
	return ret0
}

// MountResources indicates an expected call of MountResources.
func (mr *MockIContainerMockRecorder) MountResources() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MountResources", reflect.TypeOf((*MockIContainer)(nil).MountResources))
}

// PostRun mocks base method.
func (m *MockIContainer) PostRun(config *configuration.Configuration, dnsCache *dns.Records) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PostRun", config, dnsCache)
	ret0, _ := ret[0].(error)
	return ret0
}

// PostRun indicates an expected call of PostRun.
func (mr *MockIContainerMockRecorder) PostRun(config, dnsCache any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PostRun", reflect.TypeOf((*MockIContainer)(nil).PostRun), config, dnsCache)
}

// PreRun mocks base method.
func (m *MockIContainer) PreRun(config *configuration.Configuration, client *client.Http, user *authentication.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PreRun", config, client, user)
	ret0, _ := ret[0].(error)
	return ret0
}

// PreRun indicates an expected call of PreRun.
func (mr *MockIContainerMockRecorder) PreRun(config, client, user any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PreRun", reflect.TypeOf((*MockIContainer)(nil).PreRun), config, client, user)
}

// RemoveDns mocks base method.
func (m *MockIContainer) RemoveDns(dnsCache *dns.Records, networkId string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveDns", dnsCache, networkId)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveDns indicates an expected call of RemoveDns.
func (mr *MockIContainerMockRecorder) RemoveDns(dnsCache, networkId any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveDns", reflect.TypeOf((*MockIContainer)(nil).RemoveDns), dnsCache, networkId)
}

// Rename mocks base method.
func (m *MockIContainer) Rename(newName string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Rename", newName)
	ret0, _ := ret[0].(error)
	return ret0
}

// Rename indicates an expected call of Rename.
func (mr *MockIContainerMockRecorder) Rename(newName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Rename", reflect.TypeOf((*MockIContainer)(nil).Rename), newName)
}

// Restart mocks base method.
func (m *MockIContainer) Restart() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Restart")
	ret0, _ := ret[0].(error)
	return ret0
}

// Restart indicates an expected call of Restart.
func (mr *MockIContainerMockRecorder) Restart() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Restart", reflect.TypeOf((*MockIContainer)(nil).Restart))
}

// Run mocks base method.
func (m *MockIContainer) Run() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Run")
	ret0, _ := ret[0].(error)
	return ret0
}

// Run indicates an expected call of Run.
func (mr *MockIContainerMockRecorder) Run() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockIContainer)(nil).Run))
}

// SetGhost mocks base method.
func (m *MockIContainer) SetGhost(arg0 bool) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetGhost", arg0)
}

// SetGhost indicates an expected call of SetGhost.
func (mr *MockIContainerMockRecorder) SetGhost(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetGhost", reflect.TypeOf((*MockIContainer)(nil).SetGhost), arg0)
}

// Start mocks base method.
func (m *MockIContainer) Start() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start")
	ret0, _ := ret[0].(error)
	return ret0
}

// Start indicates an expected call of Start.
func (mr *MockIContainerMockRecorder) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockIContainer)(nil).Start))
}

// Stop mocks base method.
func (m *MockIContainer) Stop(signal string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stop", signal)
	ret0, _ := ret[0].(error)
	return ret0
}

// Stop indicates an expected call of Stop.
func (mr *MockIContainerMockRecorder) Stop(signal any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockIContainer)(nil).Stop), signal)
}

// SyncNetwork mocks base method.
func (m *MockIContainer) SyncNetwork() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SyncNetwork")
	ret0, _ := ret[0].(error)
	return ret0
}

// SyncNetwork indicates an expected call of SyncNetwork.
func (mr *MockIContainerMockRecorder) SyncNetwork() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SyncNetwork", reflect.TypeOf((*MockIContainer)(nil).SyncNetwork))
}

// ToJson mocks base method.
func (m *MockIContainer) ToJson() ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ToJson")
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ToJson indicates an expected call of ToJson.
func (mr *MockIContainerMockRecorder) ToJson() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ToJson", reflect.TypeOf((*MockIContainer)(nil).ToJson))
}

// UpdateDns mocks base method.
func (m *MockIContainer) UpdateDns(dnsCache *dns.Records) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateDns", dnsCache)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateDns indicates an expected call of UpdateDns.
func (mr *MockIContainerMockRecorder) UpdateDns(dnsCache any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateDns", reflect.TypeOf((*MockIContainer)(nil).UpdateDns), dnsCache)
}

// MockIPlatform is a mock of IPlatform interface.
type MockIPlatform struct {
	ctrl     *gomock.Controller
	recorder *MockIPlatformMockRecorder
	isgomock struct{}
}

// MockIPlatformMockRecorder is the mock recorder for MockIPlatform.
type MockIPlatformMockRecorder struct {
	mock *MockIPlatform
}

// NewMockIPlatform creates a new mock instance.
func NewMockIPlatform(ctrl *gomock.Controller) *MockIPlatform {
	mock := &MockIPlatform{ctrl: ctrl}
	mock.recorder = &MockIPlatformMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIPlatform) EXPECT() *MockIPlatformMockRecorder {
	return m.recorder
}

// Delete mocks base method.
func (m *MockIPlatform) Delete() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete")
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockIPlatformMockRecorder) Delete() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockIPlatform)(nil).Delete))
}

// Exec mocks base method.
func (m *MockIPlatform) Exec(command []string) (types.ExecResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Exec", command)
	ret0, _ := ret[0].(types.ExecResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exec indicates an expected call of Exec.
func (mr *MockIPlatformMockRecorder) Exec(command any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exec", reflect.TypeOf((*MockIPlatform)(nil).Exec), command)
}

// GetDefinition mocks base method.
func (m *MockIPlatform) GetDefinition() contracts.IDefinition {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDefinition")
	ret0, _ := ret[0].(contracts.IDefinition)
	return ret0
}

// GetDefinition indicates an expected call of GetDefinition.
func (mr *MockIPlatformMockRecorder) GetDefinition() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDefinition", reflect.TypeOf((*MockIPlatform)(nil).GetDefinition))
}

// GetDomain mocks base method.
func (m *MockIPlatform) GetDomain(networkName string) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDomain", networkName)
	ret0, _ := ret[0].(string)
	return ret0
}

// GetDomain indicates an expected call of GetDomain.
func (mr *MockIPlatformMockRecorder) GetDomain(networkName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDomain", reflect.TypeOf((*MockIPlatform)(nil).GetDomain), networkName)
}

// GetGeneratedName mocks base method.
func (m *MockIPlatform) GetGeneratedName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGeneratedName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetGeneratedName indicates an expected call of GetGeneratedName.
func (mr *MockIPlatformMockRecorder) GetGeneratedName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGeneratedName", reflect.TypeOf((*MockIPlatform)(nil).GetGeneratedName))
}

// GetGroup mocks base method.
func (m *MockIPlatform) GetGroup() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGroup")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetGroup indicates an expected call of GetGroup.
func (mr *MockIPlatformMockRecorder) GetGroup() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGroup", reflect.TypeOf((*MockIPlatform)(nil).GetGroup))
}

// GetGroupIdentifier mocks base method.
func (m *MockIPlatform) GetGroupIdentifier() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGroupIdentifier")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetGroupIdentifier indicates an expected call of GetGroupIdentifier.
func (mr *MockIPlatformMockRecorder) GetGroupIdentifier() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGroupIdentifier", reflect.TypeOf((*MockIPlatform)(nil).GetGroupIdentifier))
}

// GetHeadlessDomain mocks base method.
func (m *MockIPlatform) GetHeadlessDomain(networkName string) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetHeadlessDomain", networkName)
	ret0, _ := ret[0].(string)
	return ret0
}

// GetHeadlessDomain indicates an expected call of GetHeadlessDomain.
func (mr *MockIPlatformMockRecorder) GetHeadlessDomain(networkName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHeadlessDomain", reflect.TypeOf((*MockIPlatform)(nil).GetHeadlessDomain), networkName)
}

// GetId mocks base method.
func (m *MockIPlatform) GetId() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetId")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetId indicates an expected call of GetId.
func (mr *MockIPlatformMockRecorder) GetId() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetId", reflect.TypeOf((*MockIPlatform)(nil).GetId))
}

// GetInit mocks base method.
func (m *MockIPlatform) GetInit() platforms.IPlatform {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInit")
	ret0, _ := ret[0].(platforms.IPlatform)
	return ret0
}

// GetInit indicates an expected call of GetInit.
func (mr *MockIPlatformMockRecorder) GetInit() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInit", reflect.TypeOf((*MockIPlatform)(nil).GetInit))
}

// GetInitDefinition mocks base method.
func (m *MockIPlatform) GetInitDefinition() v1.ContainersInternal {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInitDefinition")
	ret0, _ := ret[0].(v1.ContainersInternal)
	return ret0
}

// GetInitDefinition indicates an expected call of GetInitDefinition.
func (mr *MockIPlatformMockRecorder) GetInitDefinition() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInitDefinition", reflect.TypeOf((*MockIPlatform)(nil).GetInitDefinition))
}

// GetName mocks base method.
func (m *MockIPlatform) GetName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetName indicates an expected call of GetName.
func (mr *MockIPlatformMockRecorder) GetName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetName", reflect.TypeOf((*MockIPlatform)(nil).GetName))
}

// GetState mocks base method.
func (m *MockIPlatform) GetState() (state.State, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetState")
	ret0, _ := ret[0].(state.State)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetState indicates an expected call of GetState.
func (mr *MockIPlatformMockRecorder) GetState() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetState", reflect.TypeOf((*MockIPlatform)(nil).GetState))
}

// InitContainer mocks base method.
func (m *MockIPlatform) InitContainer(definition v1.ContainersInternal, config *configuration.Configuration, client *client.Http, user *authentication.User, runtime *types.Runtime) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InitContainer", definition, config, client, user, runtime)
	ret0, _ := ret[0].(error)
	return ret0
}

// InitContainer indicates an expected call of InitContainer.
func (mr *MockIPlatformMockRecorder) InitContainer(definition, config, client, user, runtime any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InitContainer", reflect.TypeOf((*MockIPlatform)(nil).InitContainer), definition, config, client, user, runtime)
}

// Kill mocks base method.
func (m *MockIPlatform) Kill(signal string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Kill", signal)
	ret0, _ := ret[0].(error)
	return ret0
}

// Kill indicates an expected call of Kill.
func (mr *MockIPlatformMockRecorder) Kill(signal any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Kill", reflect.TypeOf((*MockIPlatform)(nil).Kill), signal)
}

// Logs mocks base method.
func (m *MockIPlatform) Logs(arg0 bool) (io.ReadCloser, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Logs", arg0)
	ret0, _ := ret[0].(io.ReadCloser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Logs indicates an expected call of Logs.
func (mr *MockIPlatformMockRecorder) Logs(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Logs", reflect.TypeOf((*MockIPlatform)(nil).Logs), arg0)
}

// MountResources mocks base method.
func (m *MockIPlatform) MountResources() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MountResources")
	ret0, _ := ret[0].(error)
	return ret0
}

// MountResources indicates an expected call of MountResources.
func (mr *MockIPlatformMockRecorder) MountResources() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MountResources", reflect.TypeOf((*MockIPlatform)(nil).MountResources))
}

// PostRun mocks base method.
func (m *MockIPlatform) PostRun(config *configuration.Configuration, dnsCache *dns.Records) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PostRun", config, dnsCache)
	ret0, _ := ret[0].(error)
	return ret0
}

// PostRun indicates an expected call of PostRun.
func (mr *MockIPlatformMockRecorder) PostRun(config, dnsCache any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PostRun", reflect.TypeOf((*MockIPlatform)(nil).PostRun), config, dnsCache)
}

// PreRun mocks base method.
func (m *MockIPlatform) PreRun(config *configuration.Configuration, client *client.Http, user *authentication.User, runtime *types.Runtime) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PreRun", config, client, user, runtime)
	ret0, _ := ret[0].(error)
	return ret0
}

// PreRun indicates an expected call of PreRun.
func (mr *MockIPlatformMockRecorder) PreRun(config, client, user, runtime any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PreRun", reflect.TypeOf((*MockIPlatform)(nil).PreRun), config, client, user, runtime)
}

// RemoveDns mocks base method.
func (m *MockIPlatform) RemoveDns(dnsCache *dns.Records, networkId string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveDns", dnsCache, networkId)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveDns indicates an expected call of RemoveDns.
func (mr *MockIPlatformMockRecorder) RemoveDns(dnsCache, networkId any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveDns", reflect.TypeOf((*MockIPlatform)(nil).RemoveDns), dnsCache, networkId)
}

// Rename mocks base method.
func (m *MockIPlatform) Rename(newName string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Rename", newName)
	ret0, _ := ret[0].(error)
	return ret0
}

// Rename indicates an expected call of Rename.
func (mr *MockIPlatformMockRecorder) Rename(newName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Rename", reflect.TypeOf((*MockIPlatform)(nil).Rename), newName)
}

// Restart mocks base method.
func (m *MockIPlatform) Restart() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Restart")
	ret0, _ := ret[0].(error)
	return ret0
}

// Restart indicates an expected call of Restart.
func (mr *MockIPlatformMockRecorder) Restart() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Restart", reflect.TypeOf((*MockIPlatform)(nil).Restart))
}

// Run mocks base method.
func (m *MockIPlatform) Run() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Run")
	ret0, _ := ret[0].(error)
	return ret0
}

// Run indicates an expected call of Run.
func (mr *MockIPlatformMockRecorder) Run() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockIPlatform)(nil).Run))
}

// Start mocks base method.
func (m *MockIPlatform) Start() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start")
	ret0, _ := ret[0].(error)
	return ret0
}

// Start indicates an expected call of Start.
func (mr *MockIPlatformMockRecorder) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockIPlatform)(nil).Start))
}

// Stop mocks base method.
func (m *MockIPlatform) Stop(signal string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stop", signal)
	ret0, _ := ret[0].(error)
	return ret0
}

// Stop indicates an expected call of Stop.
func (mr *MockIPlatformMockRecorder) Stop(signal any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockIPlatform)(nil).Stop), signal)
}

// SyncNetwork mocks base method.
func (m *MockIPlatform) SyncNetwork() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SyncNetwork")
	ret0, _ := ret[0].(error)
	return ret0
}

// SyncNetwork indicates an expected call of SyncNetwork.
func (mr *MockIPlatformMockRecorder) SyncNetwork() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SyncNetwork", reflect.TypeOf((*MockIPlatform)(nil).SyncNetwork))
}

// ToJson mocks base method.
func (m *MockIPlatform) ToJson() ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ToJson")
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ToJson indicates an expected call of ToJson.
func (mr *MockIPlatformMockRecorder) ToJson() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ToJson", reflect.TypeOf((*MockIPlatform)(nil).ToJson))
}

// UpdateDns mocks base method.
func (m *MockIPlatform) UpdateDns(dnsCache *dns.Records) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateDns", dnsCache)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateDns indicates an expected call of UpdateDns.
func (mr *MockIPlatformMockRecorder) UpdateDns(dnsCache any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateDns", reflect.TypeOf((*MockIPlatform)(nil).UpdateDns), dnsCache)
}

// MockRegistry is a mock of Registry interface.
type MockRegistry struct {
	ctrl     *gomock.Controller
	recorder *MockRegistryMockRecorder
	isgomock struct{}
}

// MockRegistryMockRecorder is the mock recorder for MockRegistry.
type MockRegistryMockRecorder struct {
	mock *MockRegistry
}

// NewMockRegistry creates a new mock instance.
func NewMockRegistry(ctrl *gomock.Controller) *MockRegistry {
	mock := &MockRegistry{ctrl: ctrl}
	mock.recorder = &MockRegistryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRegistry) EXPECT() *MockRegistryMockRecorder {
	return m.recorder
}

// AddOrUpdate mocks base method.
func (m *MockRegistry) AddOrUpdate(group, name string, containerAddr platforms.IContainer) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddOrUpdate", group, name, containerAddr)
}

// AddOrUpdate indicates an expected call of AddOrUpdate.
func (mr *MockRegistryMockRecorder) AddOrUpdate(group, name, containerAddr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddOrUpdate", reflect.TypeOf((*MockRegistry)(nil).AddOrUpdate), group, name, containerAddr)
}

// BackOff mocks base method.
func (m *MockRegistry) BackOff(group, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BackOff", group, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// BackOff indicates an expected call of BackOff.
func (mr *MockRegistryMockRecorder) BackOff(group, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BackOff", reflect.TypeOf((*MockRegistry)(nil).BackOff), group, name)
}

// BackOffReset mocks base method.
func (m *MockRegistry) BackOffReset(group, name string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "BackOffReset", group, name)
}

// BackOffReset indicates an expected call of BackOffReset.
func (mr *MockRegistryMockRecorder) BackOffReset(group, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BackOffReset", reflect.TypeOf((*MockRegistry)(nil).BackOffReset), group, name)
}

// Find mocks base method.
func (m *MockRegistry) Find(prefix, group, name string) platforms.IContainer {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Find", prefix, group, name)
	ret0, _ := ret[0].(platforms.IContainer)
	return ret0
}

// Find indicates an expected call of Find.
func (mr *MockRegistryMockRecorder) Find(prefix, group, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Find", reflect.TypeOf((*MockRegistry)(nil).Find), prefix, group, name)
}

// FindGroup mocks base method.
func (m *MockRegistry) FindGroup(prefix, group string) []platforms.IContainer {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindGroup", prefix, group)
	ret0, _ := ret[0].([]platforms.IContainer)
	return ret0
}

// FindGroup indicates an expected call of FindGroup.
func (mr *MockRegistryMockRecorder) FindGroup(prefix, group any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindGroup", reflect.TypeOf((*MockRegistry)(nil).FindGroup), prefix, group)
}

// FindLocal mocks base method.
func (m *MockRegistry) FindLocal(group, name string) platforms.IContainer {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindLocal", group, name)
	ret0, _ := ret[0].(platforms.IContainer)
	return ret0
}

// FindLocal indicates an expected call of FindLocal.
func (mr *MockRegistryMockRecorder) FindLocal(group, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindLocal", reflect.TypeOf((*MockRegistry)(nil).FindLocal), group, name)
}

// GetIndexes mocks base method.
func (m *MockRegistry) GetIndexes(group, name string) []uint64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetIndexes", group, name)
	ret0, _ := ret[0].([]uint64)
	return ret0
}

// GetIndexes indicates an expected call of GetIndexes.
func (mr *MockRegistryMockRecorder) GetIndexes(group, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetIndexes", reflect.TypeOf((*MockRegistry)(nil).GetIndexes), group, name)
}

// Name mocks base method.
func (m *MockRegistry) Name(client *client.Http, group, name string) (string, []uint64) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Name", client, group, name)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].([]uint64)
	return ret0, ret1
}

// Name indicates an expected call of Name.
func (mr *MockRegistryMockRecorder) Name(client, group, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Name", reflect.TypeOf((*MockRegistry)(nil).Name), client, group, name)
}

// NameReplica mocks base method.
func (m *MockRegistry) NameReplica(group, name string, index uint64) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NameReplica", group, name, index)
	ret0, _ := ret[0].(string)
	return ret0
}

// NameReplica indicates an expected call of NameReplica.
func (mr *MockRegistryMockRecorder) NameReplica(group, name, index any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NameReplica", reflect.TypeOf((*MockRegistry)(nil).NameReplica), group, name, index)
}

// Remove mocks base method.
func (m *MockRegistry) Remove(prefix, group, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Remove", prefix, group, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// Remove indicates an expected call of Remove.
func (mr *MockRegistryMockRecorder) Remove(prefix, group, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Remove", reflect.TypeOf((*MockRegistry)(nil).Remove), prefix, group, name)
}

// Sync mocks base method.
func (m *MockRegistry) Sync(group, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Sync", group, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// Sync indicates an expected call of Sync.
func (mr *MockRegistryMockRecorder) Sync(group, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Sync", reflect.TypeOf((*MockRegistry)(nil).Sync), group, name)
}
