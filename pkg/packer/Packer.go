package packer

import (
	"bytes"
	"encoding/json"
	"errors"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

func New() *Pack {
	return &Pack{
		Name:        "",
		Version:     "",
		Definitions: make([]*common.Request, 0),
	}
}

func Parse(bytes []byte) ([]*common.Request, error) {
	parsed, err := ParseYAML(bytes)

	if err != nil {
		return nil, err
	}

	var request *common.Request
	requests := make([]*common.Request, 0)
	definition := v1.CommonDefinition{}

	for _, jsonRaw := range parsed {
		err = definition.FromJson(jsonRaw)

		if err != nil {
			return requests, err
		}

		request, err = common.NewRequest(definition.GetKind())

		if err != nil {
			return requests, err
		}

		err = request.Definition.FromJson(jsonRaw)

		if err != nil {
			return requests, err
		}

		requests = append(requests, request)
	}

	return requests, nil
}

func ParseYAML(yamlBytes []byte) ([]json.RawMessage, error) {
	var data = make([]json.RawMessage, 0)
	var err error

	YAML := bytes.NewBuffer(yamlBytes)

	dec := yaml.NewDecoder(YAML)

	for {
		var element map[string]interface{}
		err = dec.Decode(&element)

		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, err
		}

		tmp, err := json.Marshal(element)

		if err != nil {
			return nil, err
		}

		data = append(data, tmp)
	}

	return data, nil
}

func ReadYAMLFile(path string) ([]byte, error) {
	YAML, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	return YAML, nil
}
