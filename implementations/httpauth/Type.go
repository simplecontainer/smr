package main

import "github.com/qdnqn/smr/implementations/httpauth/shared"

type Implementation struct {
	Started bool
	Shared  *shared.Shared
	State   int
}

// Local contracts

const KIND string = "httpauth"
