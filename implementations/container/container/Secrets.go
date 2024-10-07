package container

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/template"
)

func UnpackSecretsEnvs(client *client.Http, user *authentication.User, envs []string) ([]string, error) {
	envsParsed := make([]string, 0)
	obj := objects.New(client.Get(user.Username), user)

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

func UnpackSecretsResources(client *client.Http, user *authentication.User, resource string) string {
	obj := objects.New(client.Get(user.Username), user)
	resourceParsed, err := template.ParseSecretTemplate(obj, resource)

	if err != nil {
		logger.Log.Error(err.Error())
	}

	return resourceParsed
}

func UnpackSecretsReadiness(client *client.Http, user *authentication.User, body map[string]string) (map[string]string, error) {
	bodyParsed := make(map[string]string, 0)
	obj := objects.New(client.Get(user.Username), user)

	for k, v := range body {
		parsed, err := template.ParseSecretTemplate(obj, v)

		if err != nil {
			return nil, err
		}

		bodyParsed[k] = parsed
	}

	return bodyParsed, nil
}
