package main

type Operator struct{}

const KIND string = "containers"

var invalidOperators []string = []string{
	"Run",
	"ListSupported",
}
