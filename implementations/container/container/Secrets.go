package container

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/qdnqn/smr/pkg/logger"
	"github.com/qdnqn/smr/pkg/template"
)

func (container *Container) UnpackSecretsEnvs(dbEncrypted *badger.DB, envs []string) []string {
	envsParsed := make([]string, 0)

	for _, v := range envs {
		parsed, err := template.ParseSecretTemplate(dbEncrypted, v)

		if err != nil {
			logger.Log.Error(err.Error())
		}

		envsParsed = append(envsParsed, parsed)
	}

	return envsParsed
}

func (container *Container) UnpackSecretsResources(dbEncrypted *badger.DB, resource string) string {
	resourceParsed, err := template.ParseSecretTemplate(dbEncrypted, resource)

	if err != nil {
		logger.Log.Error(err.Error())
	}

	return resourceParsed
}

func (container *Container) UnpackSecretsReadiness(dbEncrypted *badger.DB, body map[string]string) map[string]string {
	bodyParsed := make(map[string]string, 0)

	for k, v := range body {
		parsed, err := template.ParseSecretTemplate(dbEncrypted, v)

		if err != nil {
			logger.Log.Error(err.Error())
		}

		bodyParsed[k] = parsed
	}

	return bodyParsed
}
