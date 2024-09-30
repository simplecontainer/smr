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
		out, err = yaml.Marshal(v1.CertKey{})
		break
	case "configuration":
		out, err = yaml.Marshal(v1.Configuration{})
		break
	case "resource":
		out, err = yaml.Marshal(v1.Resource{})
		break
	case "httpauth":
		out, err = yaml.Marshal(v1.HttpAuth{})
		break
	case "containers":
		out, err = yaml.Marshal(v1.Containers{})
		break
	case "container":
		out, err = yaml.Marshal(v1.Container{})
		break
	case "gitops":
		out, err = yaml.Marshal(v1.Gitops{})
		break
	}

	return out, err
}
