package distributed

import (
	"github.com/simplecontainer/smr/pkg/contracts/ievents"
	"sync"
)

func NewInformer() *Informer {
	return &Informer{
		Chs:  make(map[string]map[string]chan ievents.Event),
		Lock: new(sync.RWMutex),
	}
}

func (i *Informer) AddCh(format string, event string) {
	i.Lock.Lock()
	defer i.Lock.Unlock()

	if _, ok := i.Chs[event]; !ok {
		i.Chs[event] = make(map[string]chan ievents.Event)
	}

	i.Chs[event][format] = make(chan ievents.Event)
}

func (i *Informer) GetCh(format string, event string) chan ievents.Event {
	i.Lock.RLock()
	defer i.Lock.RUnlock()

	ch, ok := i.Chs[event][format]

	if ok {
		return ch
	} else {
		return nil
	}
}

func (i *Informer) RmCh(format string, event string) {
	i.Lock.Lock()
	defer i.Lock.Unlock()

	ch, ok := i.Chs[format][event]

	if ok {
		delete(i.Chs[event], format)
		close(ch)
	}
}
