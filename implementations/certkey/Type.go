package main

import "github.com/qdnqn/smr/implementations/certkey/shared"

type Implementation struct {
	Started bool
	Shared  *shared.Shared
	State   int
}

// Local contracts

const KIND string = "certkey"
