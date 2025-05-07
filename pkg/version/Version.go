package version

import "strings"

func New(image string, version string) *Version {
	return &Version{
		Image: strings.TrimSpace(image),
		Node:  strings.TrimSpace(version),
	}
}

func NewClient(version string) *VersionClient {
	return &VersionClient{
		Version: strings.TrimSpace(version),
	}
}
