package certkey

type Certkey struct{}

const KIND string = "certkey"

var invalidOperators []string = []string{
	"Run",
	"ListSupported",
}
