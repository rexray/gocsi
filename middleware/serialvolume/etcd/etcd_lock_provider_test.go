package etcd_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	csietcd "github.com/dell/gocsi/middleware/serialvolume/etcd"
	mwtypes "github.com/dell/gocsi/middleware/serialvolume/types"
)

var p mwtypes.VolumeLockerProvider

func TestMain(m *testing.M) {
	log.SetLevel(log.InfoLevel)
	if os.Getenv(csietcd.EnvVarEndpoints) == "" {
		os.Exit(0)
	}
	os.Setenv(csietcd.EnvVarDialTimeout, "1s")
	var err error
	p, err = csietcd.New(context.TODO(), "/gocsi/etcd", 0, nil)
	if err != nil {
		log.Fatalln(err)
	}
	exitCode := m.Run()
	p.(io.Closer).Close()
	os.Exit(exitCode)
}

func TestTryMutex_Lock(t *testing.T) {

	var (
		i     int
		id    = t.Name()
		wait  sync.WaitGroup
		ready = make(chan struct{}, 5)
	)

	// Wait for the goroutines with the other mutexes to finish, otherwise
	// those mutexes won't unlock and close their concurrency sessions to etcd.
	wait.Add(5)
	defer wait.Wait()

	// The context used when creating new locks and their concurrency sessions.
	ctx := context.Background()

	// The context used for the Lock functions.
	lockCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	m, err := p.GetLockWithID(ctx, id)
	if err != nil {
		t.Error(err)
		return
	}
	m.Lock()

	// Unlock m and close its session before exiting the test.
	defer m.(io.Closer).Close()
	defer m.Unlock()

	// Start five goroutines that all attempt to lock m and increment i.
	for j := 0; j < 5; j++ {
		go func() {
			defer wait.Done()

			m, err := p.GetLockWithID(ctx, id)
			if err != nil {
				t.Error(err)
				ready <- struct{}{}
				return
			}

			defer m.(io.Closer).Close()
			m.(*csietcd.TryMutex).LockCtx = lockCtx

			ready <- struct{}{}
			m.Lock()
			i++
		}()
	}

	// Give the above loop enough time to start the goroutines.
	<-ready
	time.Sleep(time.Duration(3) * time.Second)

	// Assert that i should have only been incremented once since only
	// one lock should have been obtained.
	if i > 0 {
		t.Errorf("i != 1: %d", i)
	}
}

func ExampleTryMutex_TryLock() {

	const lockName = "ExampleTryMutex_TryLock"

	// The context used when creating new locks and their concurrency sessions.
	ctx := context.Background()

	// Assign a TryMutex to m1 and then lock m1.
	m1, err := p.GetLockWithName(ctx, lockName)
	if err != nil {
		log.Error(err)
		return
	}
	defer m1.(io.Closer).Close()
	m1.Lock()

	// Start a goroutine that sleeps for one second and then
	// unlocks m1. This makes it possible for the TryLock
	// call below to lock m2.
	go func() {
		time.Sleep(time.Duration(1) * time.Second)
		m1.Unlock()
	}()

	// Try for three seconds to lock m2.
	m2, err := p.GetLockWithName(ctx, lockName)
	if err != nil {
		log.Error(err)
		return
	}
	defer m2.(io.Closer).Close()
	if m2.TryLock(time.Duration(3) * time.Second) {
		fmt.Println("lock obtained")
	}
	m2.Unlock()

	// Output: lock obtained
}

func ExampleTryMutex_TryLock_timeout() {

	const lockName = "ExampleTryMutex_TryLock_timeout"

	// The context used when creating new locks and their concurrency sessions.
	ctx := context.Background()

	// Assign a TryMutex to m1 and then lock m1.
	m1, err := p.GetLockWithName(ctx, lockName)
	if err != nil {
		log.Error(err)
		return
	}
	defer m1.(io.Closer).Close()
	defer m1.Unlock()
	m1.Lock()

	// Try for three seconds to lock m2.
	m2, err := p.GetLockWithName(ctx, lockName)
	if err != nil {
		log.Error(err)
		return
	}
	defer m2.(io.Closer).Close()
	if !m2.TryLock(time.Duration(3) * time.Second) {
		fmt.Println("lock not obtained")
	}

	// Output: lock not obtained
}
