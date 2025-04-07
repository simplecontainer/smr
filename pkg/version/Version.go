package version

import "strings"

func New(version string) *Version {
	return &Version{
		Node: strings.TrimSpace(version),
	}
}
