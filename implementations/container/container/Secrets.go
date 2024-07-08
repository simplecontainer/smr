package container

import (
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/template"
	"net/http"
)

func (container *Container) UnpackSecretsEnvs(client *http.Client, envs []string) []string {
	envsParsed := make([]string, 0)
	obj := objects.New(client)

	for _, v := range envs {
		parsed, err := template.ParseSecretTemplate(obj, v)

		if err != nil {
			logger.Log.Error(err.Error())
		}

		envsParsed = append(envsParsed, parsed)
	}

	return envsParsed
}

func (container *Container) UnpackSecretsResources(client *http.Client, resource string) string {
	obj := objects.New(client)
	resourceParsed, err := template.ParseSecretTemplate(obj, resource)

	if err != nil {
		logger.Log.Error(err.Error())
	}

	return resourceParsed
}

func (container *Container) UnpackSecretsReadiness(client *http.Client, body map[string]string) map[string]string {
	bodyParsed := make(map[string]string, 0)
	obj := objects.New(client)

	for k, v := range body {
		parsed, err := template.ParseSecretTemplate(obj, v)

		if err != nil {
			logger.Log.Error(err.Error())
		}

		bodyParsed[k] = parsed
	}

	return bodyParsed
}
