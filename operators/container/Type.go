package main

type Operator struct{}

const KIND string = "container"

var invalidOperators []string = []string{
	"Run",
	"ListSupported",
}
