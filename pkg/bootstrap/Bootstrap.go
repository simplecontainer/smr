package bootstrap

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/static"
	"os"
)

func CreateProject(node string, environment *configuration.Environment) ([]string, error) {
	if node == "" {
		return nil, errors.New("project name cannot be empty")
	}
	return CreateDirectoryTree(environment.NodeDirectory)
}

func CreateDirectoryTree(projectDir string) ([]string, error) {
	var created []string

	for _, path := range static.STRUCTURE {
		dir := fmt.Sprintf("%s/%s", projectDir, path)

		if err := os.MkdirAll(dir, 0750); err != nil {
			_ = os.RemoveAll(projectDir)
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		created = append(created, dir)
	}

	return created, nil
}
