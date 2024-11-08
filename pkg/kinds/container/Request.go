package container

import v1 "github.com/simplecontainer/smr/pkg/definitions/v1"

func NewRequest() Request {
	return Request{
		Definition: &v1.ContainerDefinition{},
	}
}
