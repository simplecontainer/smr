package config

type Config struct{}

const KIND string = "configuration"

var invalidOperators []string = []string{
	"Run",
	"ListSupported",
}
