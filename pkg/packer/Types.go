package packer

import "github.com/simplecontainer/smr/pkg/kinds/common"

const (
	PackageMetadataFile = "Pack.yaml"
)

type Definition struct {
	File       string
	Definition *common.Request
}

type Pack struct {
	Name        string
	Version     string
	Definitions []*Definition `yaml:"-"`
	Variables   []byte        `yaml:"-"`
}
