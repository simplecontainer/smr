package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
	"runtime"
)

func Editor(jsonBytes []byte) ([]byte, bool, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return nil, false, fmt.Errorf("failed to marshal to YAML: %w", err)
	}

	f, err := os.CreateTemp("", "edit-*.yaml")
	if err != nil {
		return nil, false, fmt.Errorf("failed to create temporary file: %w", err)
	}
	tempFileName := f.Name()
	defer os.Remove(tempFileName)

	if _, err := f.Write(yamlBytes); err != nil {
		f.Close()
		return nil, false, fmt.Errorf("failed to write to temporary file: %w", err)
	}

	if err := f.Close(); err != nil {
		return nil, false, fmt.Errorf("failed to close temporary file: %w", err)
	}

	editorName := os.Getenv("EDITOR")
	if editorName == "" {
		// Default editors based on OS
		if runtime.GOOS == "windows" {
			editorName = "notepad"
		} else {
			editorName = "vi" // Default for Unix-like systems
		}
	}

	cmd := exec.Command(editorName, tempFileName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, false, fmt.Errorf("editor process failed: %w", err)
	}

	newYamlBytes, err := os.ReadFile(tempFileName)
	if err != nil {
		return nil, false, fmt.Errorf("failed to read modified file: %w", err)
	}

	if bytes.Equal(newYamlBytes, yamlBytes) {
		return jsonBytes, false, nil // No changes, return original JSON
	}

	var newData map[string]interface{}
	if err := yaml.Unmarshal(newYamlBytes, &newData); err != nil {
		return nil, false, fmt.Errorf("failed to parse modified YAML: %w", err)
	}

	newJSONBytes, err := json.Marshal(newData)
	if err != nil {
		return nil, false, fmt.Errorf("failed to convert back to JSON: %w", err)
	}

	return newJSONBytes, true, nil
}
