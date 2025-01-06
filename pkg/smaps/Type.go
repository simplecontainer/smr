package smaps

import (
	"sync"
)

type Smap struct {
	Map     sync.Map
	Members int
}
