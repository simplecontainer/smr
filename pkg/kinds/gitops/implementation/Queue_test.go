package implementation

import (
	"sync"
	"testing"

	"github.com/simplecontainer/smr/pkg/f"
	"github.com/stretchr/testify/assert"
)

// ============================================================================
// UNIT TESTS: Queue Operations
// ============================================================================

func TestNewQueueTS(t *testing.T) {
	queue := NewQueueTS()

	assert.NotNil(t, queue)
	assert.NotNil(t, queue.patches)
	assert.Equal(t, 0, queue.Size())
	assert.True(t, queue.IsEmpty())
}

func TestQueueTS_Insert(t *testing.T) {
	queue := NewQueueTS()

	commit1 := &Commit{
		Format:  f.New("smr", "kind", "containers", "app", "nginx"),
		Message: "Update nginx",
	}

	queue.Insert(commit1)

	assert.Equal(t, 1, queue.Size())
	assert.False(t, queue.IsEmpty())
}

func TestQueueTS_InsertMultiple(t *testing.T) {
	queue := NewQueueTS()

	commit1 := &Commit{Format: f.New("smr", "kind", "containers", "app", "nginx"), Message: "Update 1"}
	commit2 := &Commit{Format: f.New("smr", "kind", "containers", "app", "mysql"), Message: "Update 2"}
	commit3 := &Commit{Format: f.New("smr", "kind", "containers", "app", "redis"), Message: "Update 3"}

	queue.Insert(commit1)
	queue.Insert(commit2)
	queue.Insert(commit3)

	assert.Equal(t, 3, queue.Size())
	assert.False(t, queue.IsEmpty())
}

func TestQueueTS_Pop(t *testing.T) {
	queue := NewQueueTS()

	commit1 := &Commit{Format: f.New("smr", "kind", "containers", "app", "nginx"), Message: "First"}
	commit2 := &Commit{Format: f.New("smr", "kind", "containers", "app", "mysql"), Message: "Second"}

	queue.Insert(commit1)
	queue.Insert(commit2)

	popped := queue.Pop()

	assert.NotNil(t, popped)
	assert.Equal(t, "First", popped.Message)
	assert.Equal(t, 1, queue.Size())
}

func TestQueueTS_PopEmpty(t *testing.T) {
	queue := NewQueueTS()

	popped := queue.Pop()

	assert.Nil(t, popped)
	assert.True(t, queue.IsEmpty())
}

func TestQueueTS_PopAll(t *testing.T) {
	queue := NewQueueTS()

	commit1 := &Commit{Message: "First"}
	commit2 := &Commit{Message: "Second"}

	queue.Insert(commit1)
	queue.Insert(commit2)

	first := queue.Pop()
	second := queue.Pop()
	third := queue.Pop()

	assert.Equal(t, "First", first.Message)
	assert.Equal(t, "Second", second.Message)
	assert.Nil(t, third)
	assert.True(t, queue.IsEmpty())
}

func TestQueueTS_Peek(t *testing.T) {
	queue := NewQueueTS()

	commit1 := &Commit{Message: "First"}
	commit2 := &Commit{Message: "Second"}

	queue.Insert(commit1)
	queue.Insert(commit2)

	peeked := queue.Peek()

	assert.NotNil(t, peeked)
	assert.Equal(t, "First", peeked.Message)
	assert.Equal(t, 2, queue.Size()) // Size should not change
}

func TestQueueTS_PeekEmpty(t *testing.T) {
	queue := NewQueueTS()

	peeked := queue.Peek()

	assert.Nil(t, peeked)
}

func TestQueueTS_Size(t *testing.T) {
	queue := NewQueueTS()

	assert.Equal(t, 0, queue.Size())

	queue.Insert(&Commit{Message: "1"})
	assert.Equal(t, 1, queue.Size())

	queue.Insert(&Commit{Message: "2"})
	assert.Equal(t, 2, queue.Size())

	queue.Pop()
	assert.Equal(t, 1, queue.Size())

	queue.Pop()
	assert.Equal(t, 0, queue.Size())
}

func TestQueueTS_IsEmpty(t *testing.T) {
	queue := NewQueueTS()

	assert.True(t, queue.IsEmpty())

	queue.Insert(&Commit{Message: "test"})
	assert.False(t, queue.IsEmpty())

	queue.Pop()
	assert.True(t, queue.IsEmpty())
}

func TestQueueTS_GetSnapshot(t *testing.T) {
	queue := NewQueueTS()

	commit1 := &Commit{Message: "First"}
	commit2 := &Commit{Message: "Second"}
	commit3 := &Commit{Message: "Third"}

	queue.Insert(commit1)
	queue.Insert(commit2)
	queue.Insert(commit3)

	snapshot := queue.GetSnapshot()

	assert.Equal(t, 3, len(snapshot))
	assert.Equal(t, "First", snapshot[0].Message)
	assert.Equal(t, "Second", snapshot[1].Message)
	assert.Equal(t, "Third", snapshot[2].Message)

	// Verify it's a copy - modifying snapshot shouldn't affect queue
	snapshot[0].Message = "Modified"
	assert.Equal(t, "First", queue.Peek().Message)
}

func TestQueueTS_FIFOOrder(t *testing.T) {
	queue := NewQueueTS()

	messages := []string{"First", "Second", "Third", "Fourth", "Fifth"}

	for _, msg := range messages {
		queue.Insert(&Commit{Message: msg})
	}

	for i, expected := range messages {
		popped := queue.Pop()
		assert.NotNil(t, popped, "Failed at index %d", i)
		assert.Equal(t, expected, popped.Message, "Wrong order at index %d", i)
	}

	assert.True(t, queue.IsEmpty())
}

// ============================================================================
// CONCURRENCY TESTS
// ============================================================================

func TestQueueTS_ConcurrentInsert(t *testing.T) {
	queue := NewQueueTS()
	var wg sync.WaitGroup

	numGoroutines := 100
	numInsertsPerGoroutine := 10

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numInsertsPerGoroutine; j++ {
				commit := &Commit{
					Format:  f.New("smr", "kind", "containers", "app", "test"),
					Message: "concurrent insert",
				}
				queue.Insert(commit)
			}
		}(i)
	}

	wg.Wait()

	expectedSize := numGoroutines * numInsertsPerGoroutine
	assert.Equal(t, expectedSize, queue.Size())
}

func TestQueueTS_ConcurrentPop(t *testing.T) {
	queue := NewQueueTS()

	// Pre-populate queue
	numItems := 100
	for i := 0; i < numItems; i++ {
		queue.Insert(&Commit{Message: "item"})
	}

	var wg sync.WaitGroup
	numGoroutines := 10

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numItems/numGoroutines; j++ {
				queue.Pop()
			}
		}()
	}

	wg.Wait()

	assert.True(t, queue.IsEmpty())
}

func TestQueueTS_ConcurrentMixedOperations(t *testing.T) {
	queue := NewQueueTS()
	var wg sync.WaitGroup

	numGoroutines := 50

	// Concurrent inserts
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				queue.Insert(&Commit{Message: "insert"})
			}
		}(i)
	}

	// Concurrent peeks
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				queue.Peek()
			}
		}()
	}

	// Concurrent size checks
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				queue.Size()
				queue.IsEmpty()
			}
		}()
	}

	wg.Wait()

	// Should have all inserts
	assert.Equal(t, numGoroutines*10, queue.Size())
}

func TestQueueTS_ConcurrentPopEmpty(t *testing.T) {
	queue := NewQueueTS()
	var wg sync.WaitGroup

	numGoroutines := 100

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			result := queue.Pop()
			assert.Nil(t, result)
		}()
	}

	wg.Wait()

	assert.True(t, queue.IsEmpty())
}

func TestQueueTS_ConcurrentSnapshot(t *testing.T) {
	queue := NewQueueTS()

	// Add some initial items
	for i := 0; i < 10; i++ {
		queue.Insert(&Commit{Message: "initial"})
	}

	var wg sync.WaitGroup
	snapshots := make([][]*Commit, 10)

	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func(index int) {
			defer wg.Done()
			snapshots[index] = queue.GetSnapshot()
		}(i)
	}

	wg.Wait()

	// All snapshots should have the same length
	for _, snapshot := range snapshots {
		assert.Equal(t, 10, len(snapshot))
	}
}

// ============================================================================
// EDGE CASES
// ============================================================================

func TestQueueTS_PopAfterInsertSingle(t *testing.T) {
	queue := NewQueueTS()

	commit := &Commit{Message: "single"}
	queue.Insert(commit)

	popped := queue.Pop()

	assert.Equal(t, commit, popped)
	assert.True(t, queue.IsEmpty())
}

func TestQueueTS_MultiplePopEmpty(t *testing.T) {
	queue := NewQueueTS()

	for i := 0; i < 5; i++ {
		result := queue.Pop()
		assert.Nil(t, result)
	}

	assert.True(t, queue.IsEmpty())
}

func TestQueueTS_InsertNilCommit(t *testing.T) {
	queue := NewQueueTS()

	queue.Insert(nil)

	assert.Equal(t, 1, queue.Size())

	popped := queue.Pop()
	assert.Nil(t, popped)
}

func TestQueueTS_SnapshotIndependence(t *testing.T) {
	queue := NewQueueTS()

	queue.Insert(&Commit{Message: "original"})

	snapshot1 := queue.GetSnapshot()

	queue.Insert(&Commit{Message: "new"})

	snapshot2 := queue.GetSnapshot()

	assert.Equal(t, 1, len(snapshot1))
	assert.Equal(t, 2, len(snapshot2))
	assert.Equal(t, 2, queue.Size())
}
