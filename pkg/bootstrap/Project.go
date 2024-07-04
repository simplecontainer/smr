package bootstrap

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/static"
)

func CreateProject(projectName string, configObj *configuration.Configuration) {
	if projectName == "" {
		panic("Project name cannot be empty")
	}

	projectDir := fmt.Sprintf("%s/%s/%s", configObj.Environment.HOMEDIR, static.SMR, projectName)

	CreateDirectoryTree(projectDir)
	config := GenerateConfigProject(projectDir)

	if !WriteConfiguration(config, projectDir, projectName) {
		logger.Log.Fatal("failed to create new project")
	}
}

func DeleteProject(projectName string, configObj *configuration.Configuration) {
	if projectName == "" {
		projectName = static.SMR
	}

	projectDir := fmt.Sprintf("%s/%s/%s", configObj.Environment.HOMEDIR, static.SMR, projectName)

	ClearDirectoryTree(projectDir)
}
