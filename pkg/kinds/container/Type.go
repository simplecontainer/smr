package container

type Container struct{}

const KIND string = "container"

var invalidOperators []string = []string{
	"Run",
	"ListSupported",
}
