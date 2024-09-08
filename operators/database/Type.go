package main

type Operator struct{}

const KIND string = "database"

var invalidOperators []string = []string{
	"Run",
	"ListSupported",
}
