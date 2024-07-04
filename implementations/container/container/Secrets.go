package container

import (
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/template"
	"net/http"
)

func (container *Container) UnpackSecretsEnvs(client *http.Client, envs []string) []string {
	envsParsed := make([]string, 0)

	for _, v := range envs {
		parsed, err := template.ParseSecretTemplate(client, v)

		if err != nil {
			logger.Log.Error(err.Error())
		}

		envsParsed = append(envsParsed, parsed)
	}

	return envsParsed
}

func (container *Container) UnpackSecretsResources(client *http.Client, resource string) string {
	resourceParsed, err := template.ParseSecretTemplate(client, resource)

	if err != nil {
		logger.Log.Error(err.Error())
	}

	return resourceParsed
}

func (container *Container) UnpackSecretsReadiness(client *http.Client, body map[string]string) map[string]string {
	bodyParsed := make(map[string]string, 0)

	for k, v := range body {
		parsed, err := template.ParseSecretTemplate(client, v)

		if err != nil {
			logger.Log.Error(err.Error())
		}

		bodyParsed[k] = parsed
	}

	return bodyParsed
}
