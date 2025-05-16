package registry

import (
	"encoding/json"
	"fmt"
	"github.com/simplecontainer/smr/pkg/contracts/icontrol"
)

var commandTypes = make(map[string]func() icontrol.Command)

func RegisterCommandType(name string, constructor func() icontrol.Command) {
	commandTypes[name] = constructor
}

func CreateCommand(name string) (icontrol.Command, error) {
	constructor, exists := commandTypes[name]
	if !exists {
		return nil, fmt.Errorf("unknown command type: %s", name)
	}
	return constructor(), nil
}

func UnmarshalCommand(raw json.RawMessage) (icontrol.Command, error) {
	var peek struct {
		Name string `json:"name"`
	}

	if err := json.Unmarshal(raw, &peek); err != nil {
		return nil, err
	}

	cmd, err := CreateCommand(peek.Name)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(raw, cmd); err != nil {
		return nil, err
	}

	return cmd, nil
}
