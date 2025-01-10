package common

import "github.com/simplecontainer/smr/pkg/definitions"

type Request struct {
	Definition *definitions.Definition
	Kind       string
}
