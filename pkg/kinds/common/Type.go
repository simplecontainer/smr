package common

import (
	"github.com/simplecontainer/smr/pkg/contracts/ievents"
	"github.com/simplecontainer/smr/pkg/definitions"
)

type Request struct {
	Definition *definitions.Definition
	DeleteC    chan ievents.Event `json:"-"`
	Kind       string
	Error      error
}
