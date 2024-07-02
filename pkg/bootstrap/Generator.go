package bootstrap

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/config"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/static"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"os"
)

func CreateDirectoryTree(projectDir string) {
	for _, path := range static.STRUCTURE {
		dir := fmt.Sprintf("%s/%s", projectDir, path)

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			logger.Log.Info("Creating directory.", zap.String("Directory", dir))

			err := os.MkdirAll(dir, 0750)
			if err != nil {
				logger.Log.Fatal(err.Error())

				err := os.RemoveAll(projectDir)
				if err != nil {
					logger.Log.Fatal(err.Error())
				}
			}
		}
	}
}

func ClearDirectoryTree(projectDir string) {
	err := os.RemoveAll(projectDir)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}
}

func GenerateConfigProject(projectDir string) config.Configuration {
	return config.Configuration{
		config.Environment{
			Target: "development",
			Root:   projectDir,
		},
	}
}

func WriteConfiguration(config config.Configuration, projectDir string, configName string) bool {
	dump, err := yaml.Marshal(config)

	if err != nil {
		logger.Log.Fatal(err.Error())
		return false
	}

	err = os.WriteFile(fmt.Sprintf("%s/%s/%s.yaml", projectDir, static.CONFIGDIR, configName), dump, 0750)
	if err != nil {
		logger.Log.Fatal(err.Error())
		return false
	}

	return true
}
