package containers

type Containers struct{}

const KIND string = "containers"

var invalidOperators []string = []string{
	"Run",
	"ListSupported",
}
