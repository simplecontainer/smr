package helpers

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/contracts/iformat"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/static"
	"strings"
)

func BuildFormat(arg string, group string) (iformat.Format, error) {
	// Build proper format from arg based on info provided
	// Default to prefix=simplecontainer.io, category=kind if missing

	var format iformat.Format
	var err error = nil

	split := strings.Split(arg, "/")

	switch len(split) {
	case 1:
		// kind
		format = f.NewFromString(fmt.Sprintf("%s/%s/%s", static.SMR_PREFIX, "kind", split[0]))
		break
	case 2:
		// kind/name -> read group from flag!
		format = f.NewFromString(fmt.Sprintf("%s/%s/%s/%s/%s", static.SMR_PREFIX, "kind", split[0], group, split[1]))
		break
	case 3:
		// kind/group/name -> read group from arg
		format = f.NewFromString(fmt.Sprintf("%s/%s/%s/%s/%s", static.SMR_PREFIX, "kind", split[0], split[1], split[2]))
		break
	case 4:
		// category/kind/group/name
		format = f.NewFromString(fmt.Sprintf("%s/%s/%s/%s/%s", static.SMR_PREFIX, split[0], split[1], split[2], split[3]))
		break
	case 5:
		// prefix/category/kind/group/name
		format = f.NewFromString(fmt.Sprintf("%s/%s/%s/%s/%s", split[0], split[1], split[2], split[3], split[4]))
		break
	default:
		err = errors.New("valid formats are: [prefix/category/kind/group/name, category/kind/group/name, kind/group/name, kind/name, kind]")
	}

	return format, err
}
