package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/definitions"
	"github.com/simplecontainer/smr/pkg/f"
)

func NewRequest(kind string) (*Request, error) {
	request := &Request{
		Definition: definitions.New(kind),
	}

	if request.Definition == nil {
		return nil, errors.New(fmt.Sprintf("kind is not defined as definition %s", kind))
	}

	return request, nil
}

func (request *Request) Resolve(obj contracts.ObjectInterface, format *f.Format) error {
	err := obj.Find(format)

	if err != nil {
		return err
	}

	if obj.Exists() {
		return json.Unmarshal(obj.GetDefinitionByte(), request.Definition)
	}

	return errors.New(fmt.Sprintf("object not found: %s", format.ToString()))
}
