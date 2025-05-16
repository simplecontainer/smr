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

type LockManager struct {
	locks map[string]*os.File
	mutex sync.Mutex
}

func NewLockManager() *LockManager {
	return &LockManager{
		locks: make(map[string]*os.File),
	}
}

var globalLockManager = NewLockManager()

func AcquireLock(path string) error {
	return globalLockManager.Acquire(path)
}

func (lm *LockManager) Acquire(path string) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	if _, exists := lm.locks[path]; exists {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create lock directory: %w", err)
	}

	lockFile, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open lock file: %w", err)
	}

	err = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		lockFile.Close()
		if err == syscall.EWOULDBLOCK {
			return errors.New("lock already held by another process")
		}
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

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

	lm.locks[path] = lockFile
	return nil
}

func ReleaseLock(path string) error {
	return globalLockManager.Release(path)
}

func (lm *LockManager) Release(path string) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	lockFile, exists := lm.locks[path]
	if !exists {
		return nil
	}

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

func (lm *LockManager) ReleaseAll() {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	for path, lockFile := range lm.locks {
		syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)
		lockFile.Close()
		delete(lm.locks, path)
	}
}

type LockStatus struct {
	Exists   bool
	Locked   bool
	LockedBy int
	Stale    bool
}

func CheckLock(path string) (LockStatus, error) {
	status := LockStatus{
		Exists: false,
		Locked: false,
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return status, nil
		}
		return status, fmt.Errorf("failed to check lock file: %w", err)
	}

	status.Exists = true

	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return status, fmt.Errorf("failed to open lock file: %w", err)
	}
	defer file.Close()

	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		if err == syscall.EWOULDBLOCK {
			status.Locked = true

			if fileInfo.Size() > 0 {
				data := make([]byte, fileInfo.Size())
				if _, err := file.Read(data); err == nil {
					pidStr := strings.TrimSpace(string(data))
					if pid, err := strconv.Atoi(pidStr); err == nil {
						status.LockedBy = pid

						if err := syscall.Kill(pid, 0); err != nil {
							if err == syscall.ESRCH {
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
		syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	}

	return status, nil
}

func BreakStaleLock(path string) (bool, error) {
	status, err := CheckLock(path)
	if err != nil {
		return false, err
	}

	if !status.Exists {
		return false, nil
	}

	if status.Locked && !status.Stale {
		return false, fmt.Errorf("lock is held by active process %d", status.LockedBy)
	}

	if err := os.Remove(path); err != nil {
		return false, fmt.Errorf("failed to remove lock file: %w", err)
	}

	return true, nil
}
