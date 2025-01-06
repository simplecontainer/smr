package smaps

import "sync"

func New() *Smap {
	return &Smap{
		Map:     sync.Map{},
		Members: 0,
	}
}

func NewFromMap[K comparable, V any](m map[K]V) *Smap {
	smap := &Smap{
		Map:     sync.Map{},
		Members: 0,
	}

	for k, v := range m {
		smap.Map.Store(k, v)
		smap.Members += 1
	}

	return smap
}

func (smap *Smap) Add(k any, v any) {
	smap.Map.Store(k, v)
	smap.Members += 1
}

func (smap *Smap) Remove(k any) {
	smap.Map.Delete(k)
	members := 0

	smap.Map.Range(func(k, v interface{}) bool {
		members += 1
		return true
	})
}
