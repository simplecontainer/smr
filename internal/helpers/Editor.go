package helpers

import (
	"bytes"
	"encoding/json"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
)

func Editor(jsonBytes []byte) ([]byte, bool, error) {
	data := make(map[string]interface{})

	err := yaml.Unmarshal(jsonBytes, &data)

	if err != nil {
		return nil, false, err
	}

	yamlBytes, err := yaml.Marshal(data)

	if err != nil {
		return nil, false, err
	}

	f, err := os.CreateTemp("", "edit")

	if err != nil {
		return nil, false, err
	}

	defer os.Remove(f.Name())

	_, err = f.Write(yamlBytes)

	if err != nil {
		return nil, false, err
	}

	cmd := exec.Command("vi", f.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()

	if err != nil {
		return nil, false, err
	}

	err = cmd.Wait()

	if err != nil {
		return nil, false, err
	}

	newYamlBytes, err := os.ReadFile(f.Name()) // just pass the file name
	if err != nil {
		return nil, false, err
	}

	err = yaml.Unmarshal(newYamlBytes, &data)

	if err != nil {
		return nil, false, err
	}

	newJsonBytes, err := json.Marshal(data)

	if err != nil {
		return nil, false, err
	}

	if bytes.Equal(newYamlBytes, yamlBytes) {
		return newJsonBytes, false, nil
	} else {
		return newJsonBytes, true, nil
	}
}
