package resource

type Resource struct{}

const KIND string = "resource"

var invalidOperators []string = []string{
	"Run",
	"ListSupported",
}
