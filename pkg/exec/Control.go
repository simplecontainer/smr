package exec

import (
	"encoding/json"
	"fmt"
)

func NewResize(width, height int) (*Control, error) {
	return NewControl(RESIZE_TYPE, Resize{Width: width, Height: height})
}

func NewControl(t int, v any) (*Control, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return &Control{
		Type: t,
		Data: data,
	}, nil
}

func (c *Control) Marshal() ([]byte, error) {
	return json.Marshal(c)
}

func UnmarshalControl(data []byte) (*Control, error) {
	var c Control
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Control) DecodeResize() (*Resize, error) {
	if c.Type != RESIZE_TYPE {
		return nil, fmt.Errorf("control type is not RESIZE_TYPE")
	}
	var r Resize
	if err := json.Unmarshal(c.Data, &r); err != nil {
		return nil, err
	}
	return &r, nil
}
