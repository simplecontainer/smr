package implementation

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/simplecontainer/smr/pkg/contracts/idefinitions"
	"github.com/simplecontainer/smr/pkg/definitions"
	"github.com/simplecontainer/smr/pkg/f"
	"gopkg.in/yaml.v3"
	"os"
)

func NewCommit() *Commit {
	return &Commit{
		Format: f.New(),
		Patch:  nil,
	}
}

func (c *Commit) Parse(format string, spec string) error {
	var err error
	c.Format, err = f.Build(format, "")
	if err != nil {
		return err
	}

	if spec == "" {
		return errors.New("patch spec can't be empty")
	}

	var data interface{}
	if err := yaml.Unmarshal([]byte(spec), &data); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	c.Patch = bytes

	return nil
}

func (c *Commit) FromJson(data []byte) error {
	c.Format = f.New()
	return json.Unmarshal(data, c)
}

func (c *Commit) ToJson() ([]byte, error) {
	return json.Marshal(c)
}

func (c *Commit) UnmarshalJSON(data []byte) error {
	var temp struct {
		Format  json.RawMessage `json:"format"`
		Patch   []byte          `json:"patch"`
		Message string          `json:"message"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	var concreteFormat f.Format
	if err := json.Unmarshal(temp.Format, &concreteFormat); err != nil {
		return err
	}

	c.Format = &concreteFormat
	c.Patch = temp.Patch
	c.Message = temp.Message
	return nil
}

func (c *Commit) GenerateClone() error {
	c.Clone = definitions.New(c.Format.GetKind())

	if c.Clone.Definition == nil {
		return errors.New(fmt.Sprintf("kind is not defined as definition: %s", c.Format.GetKind()))
	}

	c.Clone.Definition.GetMeta().SetName(c.Format.GetName())
	c.Clone.Definition.GetMeta().SetGroup(c.Format.GetGroup())

	return nil
}

func (c *Commit) ApplyPatch(definition idefinitions.IDefinition) ([]byte, error) {
	bytes, err := definition.ToJSON()

	if err != nil {
		return nil, err
	}

	err = c.Clone.FromJson(bytes)
	if err != nil {
		return nil, err
	}

	err = c.Clone.PatchJSON(c.Patch)
	if err != nil {
		return nil, err
	}

	c.Clone.SetState(nil)
	c.Clone.SetRuntime(nil)

	bytes, err = c.Clone.ToYAML()
	if err != nil {
		return nil, err
	}

	if len(bytes) == 0 {
		return nil, errors.New("gitops controller doesn't allow patches that result in 0 bytes")
	}

	c.Message = fmt.Sprintf("applied patch on the definition %s/%s/%s", definition.GetKind(), definition.GetMeta().GetGroup(), definition.GetMeta().GetName())
	return bytes, nil
}

func (c *Commit) WriteFile(path string, bytes []byte) error {
	return os.WriteFile(path, bytes, 0600)
}
