package tests

import (
	"github.com/simplecontainer/smr/pkg/contracts/idefinitions"
	"github.com/simplecontainer/smr/pkg/contracts/iobjects"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
)

// MockDefinition is a lightweight test double for idefinitions.IDefinition
type MockDefinition struct {
	prefix  string
	kind    string
	meta    *commonv1.Meta
	state   *commonv1.State
	runtime *commonv1.Runtime
}

func NewMockDefinition(prefix, kind string) *MockDefinition {
	return &MockDefinition{
		prefix:  prefix,
		kind:    kind,
		meta:    &commonv1.Meta{},
		state:   &commonv1.State{},
		runtime: &commonv1.Runtime{},
	}
}

var _ idefinitions.IDefinition = (*MockDefinition)(nil) // compile-time check

// --- IDefinition methods ---

func (d *MockDefinition) FromJson([]byte) error          { return nil }
func (d *MockDefinition) SetRuntime(r *commonv1.Runtime) { d.runtime = r }
func (d *MockDefinition) GetRuntime() *commonv1.Runtime  { return d.runtime }
func (d *MockDefinition) GetPrefix() string              { return d.prefix }
func (d *MockDefinition) GetMeta() *commonv1.Meta        { return d.meta }
func (d *MockDefinition) GetState() *commonv1.State      { return d.state }
func (d *MockDefinition) SetState(s *commonv1.State)     { d.state = s }
func (d *MockDefinition) GetKind() string                { return d.kind }
func (d *MockDefinition) ResolveReferences(iobjects.ObjectInterface) ([]idefinitions.IDefinition, error) {
	return nil, nil
}
func (d *MockDefinition) ToJSON() ([]byte, error)       { return []byte("{}"), nil }
func (d *MockDefinition) ToYAML() ([]byte, error)       { return []byte("kind: MockDefinition"), nil }
func (d *MockDefinition) ToJSONString() (string, error) { return "{}", nil }
func (d *MockDefinition) Validate() (bool, error)       { return true, nil }
