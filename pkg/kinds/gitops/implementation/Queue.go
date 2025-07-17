package implementation

import (
	"sync"
)

type QueueTS struct {
	sync.Mutex
	patches []*Commit
}

func NewQueueTS() *QueueTS {
	return &QueueTS{
		Mutex:   sync.Mutex{},
		patches: make([]*Commit, 0),
	}
}

func (s *QueueTS) Insert(value *Commit) {
	s.Lock()
	defer s.Unlock()

	s.patches = append(s.patches, value)
}

func (s *QueueTS) Pop() *Commit {
	s.Lock()
	defer s.Unlock()

	if len(s.patches) == 0 {
		return nil
	}
	first := s.patches[0]
	s.patches = s.patches[1:]
	return first
}

func (s *QueueTS) Peek() *Commit {
	s.Lock()
	defer s.Unlock()

	if len(s.patches) == 0 {
		return nil
	}
	return s.patches[0]
}

func (s *QueueTS) Size() int {
	s.Lock()
	defer s.Unlock()

	return len(s.patches)
}

func (s *QueueTS) IsEmpty() bool {
	s.Lock()
	defer s.Unlock()

	return len(s.patches) == 0
}

func (s *QueueTS) GetSnapshot() []*Commit {
	s.Lock()
	defer s.Unlock()

	snapshot := make([]*Commit, len(s.patches))
	copy(snapshot, s.patches)
	return snapshot
}
