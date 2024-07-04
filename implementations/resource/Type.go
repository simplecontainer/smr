package main

import (
	"github.com/simplecontainer/smr/implementations/resource/shared"
	"net/http"
)

type Implementation struct {
	Started bool
	Shared  *shared.Shared
	State   int
	Client  *http.Client
}

// Local contracts

const KIND string = "resource"
