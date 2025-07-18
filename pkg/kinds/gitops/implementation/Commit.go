package implementation

import (
	"encoding/json"
	"fmt"
	"github.com/simplecontainer/smr/pkg/f"
	"gopkg.in/yaml.v3"
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
		Format json.RawMessage `json:"Format"`
		Patch  []byte          `json:"Patch"`
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
	return nil
}
