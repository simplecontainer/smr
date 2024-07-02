package main

import "github.com/simplecontainer/smr/implementations/certkey/shared"

type Implementation struct {
	Started bool
	Shared  *shared.Shared
	State   int
}

// Local contracts

const KIND string = "certkey"