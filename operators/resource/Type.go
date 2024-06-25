package main

type Operator struct{}

const KIND string = "resource"

var invalidOperators []string = []string{
	"Run",
	"ListSupported",
}
