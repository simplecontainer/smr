package gitops

type Gitops struct{}

const KIND string = "gitops"

var invalidOperators []string = []string{
	"Run",
	"ListSupported",
}
