package main

import (
	"github.com/simplecontainer/smr/implementations/container/shared"
)

type Implementation struct {
	Started bool
	Shared  *shared.Shared
}

// Local contracts

const KIND string = "container"
