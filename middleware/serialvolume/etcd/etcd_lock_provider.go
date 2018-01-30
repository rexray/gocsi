package etcd

import (
	"context"
	"errors"
	"path"
	"strings"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
	etcdsync "github.com/coreos/etcd/clientv3/concurrency"
	log "github.com/sirupsen/logrus"
	"github.com/thecodeteam/gosync"

	csictx "github.com/thecodeteam/gocsi/context"
	mwtypes "github.com/thecodeteam/gocsi/middleware/serialvolume/types"
)

const (
	// EnvVarDomain is the name of the environment variable that defines
	// the lock provider's concurrency domain.
	EnvVarDomain = "X_CSI_SERIAL_VOL_ACCESS_ETCD_DOMAIN"

	// EnvVarEndpoints is the name of the environment variable that defines
	// the lock provider's etcd endoints.
	EnvVarEndpoints = "X_CSI_SERIAL_VOL_ACCESS_ETCD_ENDPOINTS"
)

var (
	// ErrNoEndpoints is returns from New when no endpoints are defined.
	ErrNoEndpoints = errors.New("no endpoints")
)

// NewConfig returns a new etcd config object.
func NewConfig() etcd.Config {
	return etcd.Config{}
}

// New returns a new etcd volume lock provider.
func New(
	ctx context.Context,
	domain string,
	config etcd.Config) (mwtypes.VolumeLockerProvider, error) {

	if domain == "" {
		domain = csictx.Getenv(ctx, EnvVarDomain)
	}
	domain = path.Join("/", domain)

	if len(config.Endpoints) == 0 {
		if val, ok := csictx.LookupEnv(ctx, EnvVarEndpoints); ok {
			if endpoints := strings.Split(val, ","); len(endpoints) > 0 {
				config.Endpoints = endpoints
			}
		}
	}

	if len(config.Endpoints) == 0 {
		return nil, ErrNoEndpoints
	}

	client, err := etcd.New(config)
	if err != nil {
		return nil, err
	}

	return &provider{client: client, domain: domain}, nil
}

type provider struct {
	client *etcd.Client
	domain string
}

func (p *provider) Close() error {
	return p.client.Close()
}

func (p *provider) GetLockWithID(
	ctx context.Context, id string) (gosync.TryLocker, error) {

	return p.getLock(ctx, path.Join(p.domain, "volumesByID", id))
}

func (p *provider) GetLockWithName(
	ctx context.Context, name string) (gosync.TryLocker, error) {

	return p.getLock(ctx, path.Join(p.domain, "volumesByName", name))
}

func (p *provider) getLock(
	ctx context.Context, pfx string) (gosync.TryLocker, error) {

	log.Debugf("EtcdVolumeLockProvider: getLock: pfx=%v", pfx)
	sess, err := etcdsync.NewSession(p.client, etcdsync.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return &TryMutex{
		ctx: ctx, sess: sess, mtx: etcdsync.NewMutex(sess, pfx)}, nil
}

// TryMutex is a mutual exclusion lock backed by etcd that implements the
// TryLocker interface.
// The zero value for a TryMutex is an unlocked mutex.
//
// A TryMutex may be copied after first use.
type TryMutex struct {
	ctx  context.Context
	sess *etcdsync.Session
	mtx  *etcdsync.Mutex

	// LockCtx, when non-nil, is the context used with Lock.
	LockCtx context.Context

	// UnlockCtx, when non-nil, is the context used with Unlock.
	UnlockCtx context.Context

	// TryLockCtx, when non-nil, is the context used with TryLock.
	TryLockCtx context.Context
}

// Lock locks m. If the lock is already in use, the calling goroutine blocks
// until the mutex is available.
func (m *TryMutex) Lock() {
	//log.Debug("TryMutex: lock")
	ctx := m.LockCtx
	if ctx == nil {
		ctx = m.ctx
	}
	if err := m.mtx.Lock(ctx); err != nil {
		log.Errorf("TryMutex: lock err: %v", err)
		if err != context.Canceled && err != context.DeadlineExceeded {
			log.Panicf("TryMutex: lock panic: %v", err)
		}
	}
}

// Unlock unlocks m. It is a run-time error if m is not locked on entry to
// Unlock.
//
// A locked TryMutex is not associated with a particular goroutine. It is
// allowed for one goroutine to lock a Mutex and then arrange for another
// goroutine to unlock it.
func (m *TryMutex) Unlock() {
	//log.Debug("TryMutex: unlock")
	ctx := m.UnlockCtx
	if ctx == nil {
		ctx = m.ctx
	}
	if err := m.mtx.Unlock(ctx); err != nil {
		log.Errorf("TryMutex: unlock err: %v", err)
		if err != context.Canceled && err != context.DeadlineExceeded {
			log.Panicf("TryMutex: unlock panic: %v", err)
		}
	}
}

// Close closes and cleans up the underlying concurrency session.
func (m *TryMutex) Close() error {
	//log.Debug("TryMutex: close")
	if err := m.sess.Close(); err != nil {
		log.Errorf("TryMutex: close err: %v", err)
		return err
	}
	return nil
}

// TryLock attempts to lock m. If no lock can be obtained in the specified
// duration then a false value is returned.
func (m *TryMutex) TryLock(timeout time.Duration) bool {

	ctx := m.TryLockCtx
	if ctx == nil {
		ctx = m.ctx
	}

	// Create a timeout context only if the timeout is greater than zero.
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	if err := m.mtx.Lock(ctx); err != nil {
		log.Errorf("TryMutex: TryLock err: %v", err)
		if err != context.Canceled && err != context.DeadlineExceeded {
			log.Panicf("TryMutex: TryLock panic: %v", err)
		}
		return false
	}
	return true
}
