package gosync_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/thecodeteam/gosync"
)

func TestTryMutex_PanicOnUnlockOfUnlockedMutex(t *testing.T) {
	defer func() {
		e := recover()
		if e != "gosync: unlock of unlocked mutex" {
			panic(e)
		}
	}()
	var m gosync.TryMutex
	m.Unlock()
}

func TestTryMutex_Lock(t *testing.T) {

	var (
		i int
		m gosync.TryMutex
	)

	// Start five goroutines that all attempt to lock m and increment i.
	for j := 0; j < 5; j++ {
		go func() {
			m.Lock()
			i++
		}()
	}

	// Give the above loop enough time to start the goroutines.
	time.Sleep(time.Duration(3) * time.Second)

	// Assert that i should have only been incremented once since only
	// one lock should have been obtained.
	if i != 1 {
		t.Fatalf("i != 1: %d", i)
	}

	// Unlock the mutex without panic.
	m.Unlock()
}

func ExampleTryMutex_TryLock() {

	// Assign a gosync.TryMutex to m and then lock it
	var m gosync.TryMutex
	m.Lock()

	// Start a goroutine that sleeps for one second and then
	// unlocks m. This makes it possible for the above TryLock
	// call to lock m.
	go func() {
		time.Sleep(time.Duration(1) * time.Second)
		m.Unlock()
	}()

	// Try for three seconds to lock m.
	if m.TryLock(time.Duration(3) * time.Second) {
		fmt.Println("lock obtained")
	}

	// Output: lock obtained
}

func ExampleTryMutex_TryLock_timeout() {

	// Assign a gosync.TryMutex to m and then lock m.
	var m gosync.TryMutex
	m.Lock()

	// Try for three seconds to lock m.
	if !m.TryLock(time.Duration(3) * time.Second) {
		fmt.Println("lock not obtained")
	}

	// Output: lock not obtained
}
