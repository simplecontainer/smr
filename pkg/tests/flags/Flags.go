package flags

import "flag"

var (
	Image     string
	Tag       string
	BinaryDir string
	Cleanup   bool
	Timeout   int
)

func init() {
	flag.StringVar(&Image, "image", "smr", "SMR image name to use")
	flag.StringVar(&Tag, "tag", "latest", "SMR image tag to use")
	flag.IntVar(&Timeout, "timeout", 60, "Timeout in seconds for operations")
}
