package main

type Operator struct{}

const KIND string = "httpauth"

var invalidOperators []string = []string{
	"Run",
	"ListSupported",
}
