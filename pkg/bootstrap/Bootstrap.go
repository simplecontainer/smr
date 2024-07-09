package bootstrap

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/static"
	"os"
)

func CreateProject(project string, configObj *configuration.Configuration) ([]string, error) {
	if project == "" {
		return nil, errors.New("project name cannot be empty")
	}

	return CreateDirectoryTree(fmt.Sprintf("%s/%s/%s", configObj.Environment.HOMEDIR, static.SMR, project))
}

func DeleteProject(project string, configObj *configuration.Configuration) error {
	if project == "" {
		return errors.New("project name cannot be empty")
	}

	projectDir := fmt.Sprintf("%s/%s/%s", configObj.Environment.HOMEDIR, static.SMR, project)

	return ClearDirectoryTree(projectDir)
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

func ClearDirectoryTree(projectDir string) error {
	err := os.RemoveAll(projectDir)

	if err != nil {
		return err
	}

	return nil
}
