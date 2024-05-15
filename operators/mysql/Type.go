package main

type Operator struct{}

const KIND string = "mysql"

var invalidOperators []string = []string{
	"Run",
	"ListSupported",
}
