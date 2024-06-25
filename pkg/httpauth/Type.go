package httpauth

import v1 "github.com/qdnqn/smr/pkg/definitions/v1"

type HttpAuth struct {
	Username   string
	Password   string
	Definition v1.HttpAuth
}
