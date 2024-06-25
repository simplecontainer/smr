package main

type Operator struct{}

const KIND string = "certkey"

var invalidOperators []string = []string{
	"Run",
	"ListSupported",
}
