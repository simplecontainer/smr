package flags

import (
	"flag"
	"github.com/simplecontainer/smr/internal/helpers"
)

var (
	Home          string
	Image         string
	Tag           string
	BinaryPath    string
	BinaryPathCli string
	ExamplesDir   string
	Timeout       int
)

func init() {
	flag.StringVar(&Home, "root", helpers.GetRealHome(), "Root directory of all file I/O - should be the home of the user")
	flag.StringVar(&BinaryPath, "binary", "smr-linux-amd64/smr", "Path to where smr binary to use - path needs to have smr binary with same name")
	flag.StringVar(&BinaryPathCli, "binaryctl", "smrctl-linux-amd64/smrctl", "Path to where smr binary to use - path needs to have smr binary with same name")
	flag.StringVar(&ExamplesDir, "examples", "../examples", "Path to where smr binary to use - path needs to have smr binary with same name")
	flag.StringVar(&Image, "image", "smr", "SMR image name to use")
	flag.StringVar(&Tag, "tag", "latest", "SMR image tag to use")
	flag.IntVar(&Timeout, "timeout", 180, "Timeout in seconds for operations")
}
