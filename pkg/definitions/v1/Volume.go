package v1

import (
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/simplecontainer/smr/pkg/contracts/idefinitions"
	"github.com/simplecontainer/smr/pkg/contracts/iobjects"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	"github.com/simplecontainer/smr/pkg/static"
)

type VolumeDefinition struct {
	Kind   string          `json:"kind" validate:"required"`
	Prefix string          `json:"prefix" validate:"required"`
	Meta   *commonv1.Meta  `json:"meta"  validate:"required"`
	Spec   *VolumeInternal `json:"spec"  validate:"required"`
	State  *commonv1.State `json:"state,omitempty"`
}

type VolumeInternal struct {
	Driver     string            `json:"driver"`
	DriverOpts map[string]string `json:"driver_opts"`
}

func NewVolume() *VolumeDefinition {
	return &VolumeDefinition{
		Kind:   "",
		Prefix: "",
		Meta: &commonv1.Meta{
			Group:   "",
			Name:    "",
			Labels:  nil,
			Runtime: &commonv1.Runtime{},
		},
		Spec:  &VolumeInternal{},
		State: nil,
	}
}

func (volume *VolumeDefinition) GetPrefix() string {
	return volume.Prefix
}

func (volume *VolumeDefinition) SetRuntime(runtime *commonv1.Runtime) {
	volume.Meta.Runtime = runtime
}

func (volume *VolumeDefinition) GetRuntime() *commonv1.Runtime {
	return volume.Meta.Runtime
}

func (volume *VolumeDefinition) GetMeta() *commonv1.Meta {
	return volume.Meta
}

func (volume *VolumeDefinition) GetState() *commonv1.State {
	return volume.State
}

func (volume *VolumeDefinition) SetState(state *commonv1.State) {
	volume.State = state
}

func (volume *VolumeDefinition) GetKind() string {
	return static.KIND_VOLUME
}

func (volume *VolumeDefinition) ResolveReferences(obj iobjects.ObjectInterface) ([]idefinitions.IDefinition, error) {
	return nil, nil
}

func (volume *VolumeDefinition) FromJson(bytes []byte) error {
	return json.Unmarshal(bytes, volume)
}

func (volume *VolumeDefinition) ToJSON() ([]byte, error) {
	bytes, err := json.Marshal(volume)
	return bytes, err
}

func (volume *VolumeDefinition) ToYAML() ([]byte, error) {
	bytes, err := json.Marshal(volume)
	return bytes, err
}

func (volume *VolumeDefinition) ToJSONString() (string, error) {
	bytes, err := json.Marshal(volume)
	return string(bytes), err
}

func (volume *VolumeDefinition) Validate() (bool, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(volume)
	if err != nil {
		var invalidValidationError *validator.InvalidValidationError
		if errors.As(err, &invalidValidationError) {
			return false, err
		}
		// from here you can create your own error messages in whatever language you wish
		return false, err
	}

	return true, nil
}
