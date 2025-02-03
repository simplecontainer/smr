package bootstrap

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/static"
	"os"
)

func CreateProject(agent string, configObj *configuration.Configuration) ([]string, error) {
	if agent == "" {
		return nil, errors.New("project name cannot be empty")
	}

	return CreateDirectoryTree(fmt.Sprintf("%s/%s", configObj.Environment.Home, agent))
}

func CreateDirectoryTree(projectDir string) ([]string, error) {
	created := []string{}
	for _, path := range static.STRUCTURE {
		dir := fmt.Sprintf("%s/%s", projectDir, path)

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0750)
			if err != nil {
				err = os.RemoveAll(projectDir)

				if err != nil {
					return nil, err
				}
			}

			created = append(created, dir)
		}
	}

	return created, nil
}
