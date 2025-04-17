package distributed

import (
	"github.com/simplecontainer/smr/pkg/contracts/ievents"
	"sync"
)

func NewInformer() *Informer {
	return &Informer{
		Chs:  make(map[string]chan ievents.Event),
		Lock: new(sync.RWMutex),
	}
}

func (i *Informer) AddCh(format string) {
	i.Lock.Lock()
	defer i.Lock.Unlock()

	i.Chs[format] = make(chan ievents.Event)
}

func (i *Informer) GetCh(format string) chan ievents.Event {
	i.Lock.RLock()
	defer i.Lock.RUnlock()

	ch, ok := i.Chs[format]

	if ok {
		return ch
	} else {
		return nil
	}
}

func (i *Informer) RmCh(format string) {
	i.Lock.Lock()
	defer i.Lock.Unlock()

	ch, ok := i.Chs[format]

	if ok {
		delete(i.Chs, format)
		close(ch)
	}
}
