package main

type Operator struct{}

const KIND string = "gitops"

var invalidOperators []string = []string{
	"Run",
	"ListSupported",
}
