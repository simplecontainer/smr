package packer

import "github.com/simplecontainer/smr/pkg/kinds/common"

type Pack struct {
	Name        string
	Version     string
	Definitions []*common.Request
	Variables   []byte
}
