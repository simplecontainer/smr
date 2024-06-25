package main

type Operator struct{}

const KIND string = "configuration"

var invalidOperators []string = []string{
	"Run",
	"ListSupported",
}
