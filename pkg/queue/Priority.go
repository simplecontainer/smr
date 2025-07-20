package queue

import (
	"container/heap"
	"context"
	"sync"
	"time"
)

type WorkType int

const (
	WorkTypeNormal WorkType = iota
	WorkTypeTicker
	WorkTypeDelete
	WorkTypeCleanup
)

const (
	PriorityCleanup = 100 // Highest - cleanup operations
	PriorityDelete  = 50  // High - delete operations
	PriorityNormal  = 25  // Medium - normal operations
	PriorityTicker  = 10  // Low - ticker operations
)

type WorkItem struct {
	Type      WorkType
	Priority  int
	Action    func()
	Timestamp time.Time
	index     int // Index in heap
}

type PriorityQueue []*WorkItem

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// Higher priority first
	if pq[i].Priority != pq[j].Priority {
		return pq[i].Priority > pq[j].Priority
	}
	// Same priority: FIFO (earlier timestamp first)
	return pq[i].Timestamp.Before(pq[j].Timestamp)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*WorkItem)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

type PriorityWorkerQueue struct {
	pq         *PriorityQueue
	mu         sync.Mutex
	notEmpty   chan struct{}
	workerPool int
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewPriorityWorkerQueue(poolSize int) *PriorityWorkerQueue {
	ctx, cancel := context.WithCancel(context.Background())
	pq := &PriorityQueue{}
	heap.Init(pq)

	return &PriorityWorkerQueue{
		pq:         pq,
		notEmpty:   make(chan struct{}, 1),
		workerPool: poolSize,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (pwq *PriorityWorkerQueue) Start() {
	for i := 0; i < pwq.workerPool; i++ {
		pwq.wg.Add(1)
		go pwq.worker()
	}
}

func (pwq *PriorityWorkerQueue) worker() {
	defer pwq.wg.Done()

	for {
		select {
		case <-pwq.ctx.Done():
			return
		case <-pwq.notEmpty:
			// Process all available items in priority order
			for {
				pwq.mu.Lock()
				if pwq.pq.Len() == 0 {
					pwq.mu.Unlock()
					break
				}

				item := heap.Pop(pwq.pq).(*WorkItem)
				pwq.mu.Unlock()

				if item.Action != nil {
					item.Action()
				}
			}
		}
	}
}

func (pwq *PriorityWorkerQueue) Submit(workType WorkType, priority int, action func()) {
	workItem := &WorkItem{
		Type:      workType,
		Priority:  priority,
		Action:    action,
		Timestamp: time.Now(),
	}

	pwq.mu.Lock()
	heap.Push(pwq.pq, workItem)
	pwq.mu.Unlock()

	// Signal that work is available
	select {
	case pwq.notEmpty <- struct{}{}:
	default:
	}
}

func (pwq *PriorityWorkerQueue) Stop() {
	pwq.cancel()
	close(pwq.notEmpty)
	pwq.wg.Wait()
}
