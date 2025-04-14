package version

import "strings"

func New(image string, version string) *Version {
	return &Version{
		Image: strings.TrimSpace(image),
		Node:  strings.TrimSpace(version),
	}
}
