package packer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/template"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func New() *Pack {
	return &Pack{
		Name:        "",
		Version:     "",
		Definitions: make([]*Definition, 0),
	}
}

func Init(name string) error {
	err := os.MkdirAll(name, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", name, err)
	}

	definitionsPath := filepath.Join(name, "definitions")
	err = os.MkdirAll(definitionsPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create definitions directory: %w", err)
	}

	pack := New()
	pack.Name = name
	pack.Version = "0.0.1"

	data, err := yaml.Marshal(&pack)
	if err != nil {
		return fmt.Errorf("failed to marshal Pack.yaml data: %w", err)
	}

	packFilePath := filepath.Join(name, "Pack.yaml")
	err = os.WriteFile(packFilePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write Pack.yaml file: %w", err)
	}

	variablesFilePath := filepath.Join(name, "definitions", "variables.yaml")
	err = os.WriteFile(variablesFilePath, nil, 0644)
	if err != nil {
		return fmt.Errorf("failed to write Pack.yaml file: %w", err)
	}

	return nil
}

func Parse(name string, bytes []byte, variables []byte, set []string) ([]*Definition, error) {
	parsed, err := ParseYAML(name, bytes, variables, set)

	if err != nil {
		return nil, err
	}

	var request *common.Request
	requests := make([]*Definition, 0)
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

		requests = append(requests, &Definition{
			File:       name,
			Definition: request,
		})
	}

	return requests, nil
}

func ParseYAML(name string, yamlBytes []byte, variablesBytes []byte, set []string) ([]json.RawMessage, error) {
	var data = make([]json.RawMessage, 0)

	v := viper.New()
	v.SetConfigType("yaml")

	if len(variablesBytes) > 0 {
		err := v.ReadConfig(bytes.NewBuffer(variablesBytes))
		if err != nil {
			return nil, err
		}
	}

	for _, item := range set {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			v.Set(key, value)
		}
	}

	values := v.AllSettings()

	var parsed string = string(yamlBytes)
	if len(values) > 0 {
		tmpl := template.New(name, string(yamlBytes), template.Variables{Values: values}, template.NewFunctionManager().FuncMap())
		var err error
		parsed, err = tmpl.Parse("{{", "}}")
		if err != nil {
			return nil, err
		}
	}

	// Parse the final YAML documents
	YAMLDefinition := bytes.NewBuffer([]byte(parsed))
	dec := yaml.NewDecoder(YAMLDefinition)

	for {
		var element map[string]interface{}
		err := dec.Decode(&element)

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
