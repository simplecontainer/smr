package container

import (
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/template"
	"net/http"
)

func (container *Container) UnpackSecretsEnvs(client *http.Client, envs []string) ([]string, error) {
	envsParsed := make([]string, 0)
	obj := objects.New(client)

	for _, v := range envs {
		parsed, err := template.ParseSecretTemplate(obj, v)

		if err != nil {
			logger.Log.Info(err.Error())
			return nil, err
		}

		envsParsed = append(envsParsed, parsed)
	}

	return envsParsed, nil
}

func (container *Container) UnpackSecretsResources(client *http.Client, resource string) string {
	obj := objects.New(client)
	resourceParsed, err := template.ParseSecretTemplate(obj, resource)

	if err != nil {
		logger.Log.Error(err.Error())
	}

	return resourceParsed
}

func UnpackSecretsReadiness(client *http.Client, body map[string]string) (map[string]string, error) {
	bodyParsed := make(map[string]string, 0)
	obj := objects.New(client)

	for k, v := range body {
		parsed, err := template.ParseSecretTemplate(obj, v)

		if err != nil {
			return nil, err
		}

		bodyParsed[k] = parsed
	}

	return bodyParsed, nil
}
