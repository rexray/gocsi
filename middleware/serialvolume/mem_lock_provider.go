package serialvolume

import (
	"context"
	"sync"

	"github.com/akutz/gosync"
)

// MemLockProvider is an in-memory implementation of LockProvider.
type MemLockProvider struct {
	sync.Once
	volIDLocksL   sync.Mutex
	volNameLocksL sync.Mutex
	volIDLocks    map[string]gosync.TryLocker
	volNameLocks  map[string]gosync.TryLocker
}

func (i *MemLockProvider) init() {
	i.volIDLocks = map[string]gosync.TryLocker{}
	i.volNameLocks = map[string]gosync.TryLocker{}
}

// GetLockWithID gets a lock for a volume with provided ID. If a lock
// for the specified volume ID does not exist then a new lock is created
// and returned.
func (i *MemLockProvider) GetLockWithID(
	ctx context.Context, id string) (gosync.TryLocker, error) {

	i.Once.Do(i.init)

	i.volIDLocksL.Lock()
	defer i.volIDLocksL.Unlock()
	lock := i.volIDLocks[id]
	if lock == nil {
		lock = &gosync.TryMutex{}
		i.volIDLocks[id] = lock
	}
	return lock, nil
}

// GetLockWithName gets a lock for a volume with provided name. If a lock
// for the specified volume name does not exist then a new lock is created
// and returned.
func (i *MemLockProvider) GetLockWithName(
	ctx context.Context, name string) (gosync.TryLocker, error) {

	i.Once.Do(i.init)

	i.volNameLocksL.Lock()
	defer i.volNameLocksL.Unlock()
	lock := i.volNameLocks[name]
	if lock == nil {
		lock = &gosync.TryMutex{}
		i.volNameLocks[name] = lock
	}
	return lock, nil
}
