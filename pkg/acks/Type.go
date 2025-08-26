package acks

import (
	"github.com/simplecontainer/smr/pkg/smaps"
	"time"
)

type Acks struct {
	Acks    *smaps.Smap
	Timeout time.Duration
}
