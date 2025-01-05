package secrets

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/secrets"
	"github.com/simplecontainer/smr/pkg/template"
)

func UnpackSecretsEnvs(client *client.Http, user *authentication.User, envs []string) ([]string, error) {
	envsParsed := make([]string, 0)
	obj := secrets.New(client.Get(user.Username), user)

	for _, v := range envs {
		parsed, err := template.ParseSecretTemplate(obj, v)

		if err != nil {
			return nil, err
		}

		envsParsed = append(envsParsed, parsed)
	}

	return envsParsed, nil
}

func UnpackSecretsResources(client *client.Http, user *authentication.User, resource string) (string, error) {
	obj := secrets.New(client.Get(user.Username), user)
	resourceParsed, err := template.ParseSecretTemplate(obj, resource)

	if err != nil {
		return resource, err
	}

	return resourceParsed, nil
}

func UnpackSecretsReadiness(client *client.Http, user *authentication.User, command []string) ([]string, error) {
	commandParsed := make([]string, 0)
	obj := secrets.New(client.Get(user.Username), user)

	for _, v := range command {
		parsed, err := template.ParseSecretTemplate(obj, v)

		if err != nil {
			return nil, err
		}

		commandParsed = append(commandParsed, parsed)
	}

	return commandParsed, nil
}
