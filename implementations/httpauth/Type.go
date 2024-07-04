package main

import (
	"github.com/simplecontainer/smr/implementations/httpauth/shared"
	"net/http"
)

type Implementation struct {
	Started bool
	Shared  *shared.Shared
	Client  *http.Client
}

// Local contracts

const KIND string = "httpauth"
