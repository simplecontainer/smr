package packer

import "github.com/simplecontainer/smr/pkg/kinds/common"

type Definition struct {
	File       string
	Definition *common.Request
}

type Pack struct {
	Name        string
	Version     string
	Definitions []*Definition
	Variables   []byte
}
