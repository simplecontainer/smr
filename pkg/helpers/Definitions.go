package helpers

import (
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"gopkg.in/yaml.v3"
)

func Definitions(defintion string) ([]byte, error) {
	var out []byte = nil
	var err error = nil

	switch defintion {
	case "certkey":
		out, err = yaml.Marshal(v1.CertKeyDefinition{})
		break
	case "configuration":
		out, err = yaml.Marshal(v1.ConfigurationDefinition{})
		break
	case "resource":
		out, err = yaml.Marshal(v1.ResourceDefinition{})
		break
	case "httpauth":
		out, err = yaml.Marshal(v1.HttpAuthDefinition{})
		break
	case "containers":
		out, err = yaml.Marshal(v1.ContainersDefinition{})
		break
	case "container":
		out, err = yaml.Marshal(v1.ContainerDefinition{})
		break
	case "gitops":
		out, err = yaml.Marshal(v1.GitopsDefinition{})
		break
	}

	return out, err
}
