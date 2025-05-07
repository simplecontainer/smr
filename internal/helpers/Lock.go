package helpers

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

// LockManager handles multiple process locks
type LockManager struct {
	locks map[string]*os.File
	mutex sync.Mutex
}

// NewLockManager creates a new lock manager instance
func NewLockManager() *LockManager {
	return &LockManager{
		locks: make(map[string]*os.File),
	}
}

// Global lock manager instance
var globalLockManager = NewLockManager()

// AcquireLock tries to acquire a lock at the specified path
// Returns nil if successful, error otherwise
func AcquireLock(path string) error {
	return globalLockManager.Acquire(path)
}

// Acquire obtains a lock with the specified name
func (lm *LockManager) Acquire(path string) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	// Check if we already hold this lock
	if _, exists := lm.locks[path]; exists {
		return nil // Already have this lock
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create lock directory: %w", err)
	}

	// Open or create lock file
	lockFile, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open lock file: %w", err)
	}

	// Try to acquire exclusive non-blocking lock
	err = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		lockFile.Close()
		if err == syscall.EWOULDBLOCK {
			return errors.New("lock already held by another process")
		}
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	// Write PID to lock file
	if err = lockFile.Truncate(0); err != nil {
		syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)
		lockFile.Close()
		return fmt.Errorf("failed to truncate lock file: %w", err)
	}

	if _, err = fmt.Fprintf(lockFile, "%d\n", os.Getpid()); err != nil {
		syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)
		lockFile.Close()
		return fmt.Errorf("failed to write PID to lock file: %w", err)
	}

	// Store lock in our map
	lm.locks[path] = lockFile
	return nil
}

// ReleaseLock releases the lock at the specified path
func ReleaseLock(path string) error {
	return globalLockManager.Release(path)
}

// Release frees a specific lock
func (lm *LockManager) Release(path string) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	lockFile, exists := lm.locks[path]
	if !exists {
		return nil // We don't hold this lock
	}

	// Release the lock
	err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)
	lockFile.Close()
	delete(lm.locks, path)

	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	return nil
}

// ReleaseAllLocks releases all locks held by this process
func ReleaseAllLocks() {
	globalLockManager.ReleaseAll()
}

// ReleaseAll frees all locks held by this manager
func (lm *LockManager) ReleaseAll() {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	for path, lockFile := range lm.locks {
		syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)
		lockFile.Close()
		delete(lm.locks, path)
	}
}

// LockStatus represents the status of a lock file
type LockStatus struct {
	Exists   bool
	Locked   bool
	LockedBy int
	Stale    bool
}

// CheckLock verifies if a lock exists and its status
func CheckLock(path string) (LockStatus, error) {
	status := LockStatus{
		Exists: false,
		Locked: false,
	}

	// Check if file exists
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return status, nil // Lock doesn't exist
		}
		return status, fmt.Errorf("failed to check lock file: %w", err)
	}

	status.Exists = true

	// Try to open the file
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return status, fmt.Errorf("failed to open lock file: %w", err)
	}
	defer file.Close()

	// Try to acquire lock to see if it's already locked
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		if err == syscall.EWOULDBLOCK {
			// File is locked by another process
			status.Locked = true

			// Read PID from lock file
			if fileInfo.Size() > 0 {
				data := make([]byte, fileInfo.Size())
				if _, err := file.Read(data); err == nil {
					pidStr := strings.TrimSpace(string(data))
					if pid, err := strconv.Atoi(pidStr); err == nil {
						status.LockedBy = pid

						// Check if process exists
						if err := syscall.Kill(pid, 0); err != nil {
							if err == syscall.ESRCH {
								// Process doesn't exist, stale lock
								status.Stale = true
							}
						}
					}
				}
			}
		} else {
			return status, fmt.Errorf("error checking lock: %w", err)
		}
	} else {
		// We got the lock, so it wasn't locked
		// Release it immediately
		syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	}

	return status, nil
}

// BreakStaleLock attempts to remove a stale lock file
// Only breaks the lock if the process that created it no longer exists
func BreakStaleLock(path string) (bool, error) {
	status, err := CheckLock(path)
	if err != nil {
		return false, err
	}

	if !status.Exists {
		return false, nil // No lock to break
	}

	if status.Locked && !status.Stale {
		return false, fmt.Errorf("lock is held by active process %d", status.LockedBy)
	}

	// If the lock exists but is not locked, or is stale, we can break it
	if err := os.Remove(path); err != nil {
		return false, fmt.Errorf("failed to remove lock file: %w", err)
	}

	return true, nil
}
