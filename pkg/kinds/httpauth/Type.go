package httpauth

type Httpauth struct{}

const KIND string = "httpauth"

var invalidOperators []string = []string{
	"Run",
	"ListSupported",
}
