package main

import "github.com/qdnqn/smr/implementations/containers/shared"

type Implementation struct {
	Started bool
	Shared  *shared.Shared
}

// Local contracts

const KIND string = "containers"
