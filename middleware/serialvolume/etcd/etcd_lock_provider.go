package etcd

import (
	"context"
	"crypto/tls"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/akutz/gosync"
	etcd "github.com/coreos/etcd/clientv3"
	etcdsync "github.com/coreos/etcd/clientv3/concurrency"
	log "github.com/sirupsen/logrus"

	csienv "github.com/rexray/gocsi/env"
)

// LockProvider is an etcd-based implementation of the
// serialvolume.LockProvider interface.
type LockProvider struct {
	sync.Once
	Domain string
	TTL    time.Duration
	Config etcd.Config
	client *etcd.Client
}

func (p *LockProvider) init(ctx context.Context) error {
	p.Domain = csienv.Getenv(ctx, EnvVarPrefix)
	p.Domain = path.Join("/", p.Domain)
	p.TTL, _ = time.ParseDuration(csienv.Getenv(ctx, EnvVarTTL))

	if v := csienv.Getenv(ctx, EnvVarEndpoints); v != "" {
		p.Config.Endpoints = strings.Split(v, ",")
	}

	if v := csienv.Getenv(ctx, EnvVarAutoSyncInterval); v != "" {
		v, err := time.ParseDuration(v)
		if err != nil {
			return err
		}
		p.Config.AutoSyncInterval = v
	}

	if v := csienv.Getenv(ctx, EnvVarDialKeepAliveTime); v != "" {
		v, err := time.ParseDuration(v)
		if err != nil {
			return err
		}
		p.Config.DialKeepAliveTime = v
	}

	if v := csienv.Getenv(ctx, EnvVarDialKeepAliveTimeout); v != "" {
		v, err := time.ParseDuration(v)
		if err != nil {
			return err
		}
		p.Config.DialKeepAliveTimeout = v
	}

	if v := csienv.Getenv(ctx, EnvVarDialTimeout); v != "" {
		v, err := time.ParseDuration(v)
		if err != nil {
			return err
		}
		p.Config.DialTimeout = v
	}

	if v := csienv.Getenv(ctx, EnvVarMaxCallRecvMsgSz); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			return err
		}
		p.Config.MaxCallRecvMsgSize = i
	}

	if v := csienv.Getenv(ctx, EnvVarMaxCallSendMsgSz); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			return err
		}
		p.Config.MaxCallSendMsgSize = i
	}

	if v := csienv.Getenv(ctx, EnvVarUsername); v != "" {
		p.Config.Username = v
	}
	if v := csienv.Getenv(ctx, EnvVarPassword); v != "" {
		p.Config.Password = v
	}

	if v, ok := csienv.LookupEnv(ctx, EnvVarRejectOldCluster); ok {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return err
		}
		p.Config.RejectOldCluster = b
	}

	if v, ok := csienv.LookupEnv(ctx, EnvVarTLS); ok {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return err
		}
		if b {
			p.Config.TLS = &tls.Config{}
			if v, ok := csienv.LookupEnv(ctx, EnvVarTLSInsecure); ok {
				b, err := strconv.ParseBool(v)
				if err != nil {
					return err
				}
				p.Config.TLS.InsecureSkipVerify = b
			}
		}
	}

	client, err := etcd.New(p.Config)
	if err != nil {
		return err
	}
	p.client = client

	log.WithField("config", p.Config).Info(
		"created etcd-enabled serial volume access lock provider")
	return nil
}

func (p *LockProvider) Close() error {
	return p.client.Close()
}

func (p *LockProvider) GetLockWithID(
	ctx context.Context, id string) (gosync.TryLocker, error) {

	var err error
	p.Once.Do(func() {
		err = p.init(ctx)
	})
	if err != nil {
		return nil, err
	}

	return p.getLock(ctx, path.Join(p.Domain, "volumesByID", id))
}

func (p *LockProvider) GetLockWithName(
	ctx context.Context, name string) (gosync.TryLocker, error) {

	var err error
	p.Once.Do(func() {
		err = p.init(ctx)
	})
	if err != nil {
		return nil, err
	}

	return p.getLock(ctx, path.Join(p.Domain, "volumesByName", name))
}

func (p *LockProvider) getLock(
	ctx context.Context, pfx string) (gosync.TryLocker, error) {

	log.Debugf("EtcdVolumeLockProvider: getLock: pfx=%v", pfx)

	opts := []etcdsync.SessionOption{etcdsync.WithContext(ctx)}
	if p.TTL > 0 {
		opts = append(opts, etcdsync.WithTTL(int(p.TTL.Seconds())))
	}

	sess, err := etcdsync.NewSession(p.client, opts...)
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
		log.Debugf("TryMutex: lock err: %v", err)
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
		log.Debugf("TryMutex: unlock err: %v", err)
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
		log.Debugf("TryMutex: TryLock err: %v", err)
		if err != context.Canceled && err != context.DeadlineExceeded {
			log.Panicf("TryMutex: TryLock panic: %v", err)
		}
		return false
	}
	return true
}

// Usage returns the lock provider's usage string.
func (i *LockProvider) Usage() string {
	return usage
}

const usage = `SERIAL VOLUME ACCESS ETCD
    X_CSI_SERIAL_VOL_ACCESS_ETCD_DOMAIN
        The name of the environment variable that defines the etcd lock
        provider's concurrency domain.

    X_CSI_SERIAL_VOL_ACCESS_ETCD_TTL
        The length of time etcd will wait before  releasing ownership of a
        distributed lock if the lock's session has not been renewed.

    X_CSI_SERIAL_VOL_ACCESS_ETCD_ENDPOINTS
        A comma-separated list of etcd endpoints. If specified then the
        SP's serial volume access middleware will leverage etcd to enable
        distributed locking.

    X_CSI_SERIAL_VOL_ACCESS_ETCD_AUTO_SYNC_INTERVAL
        A time.Duration string that specifies the interval to update
        endpoints with its latest members. A value of 0 disables
        auto-sync. By default auto-sync is disabled.

    X_CSI_SERIAL_VOL_ACCESS_ETCD_DIAL_TIMEOUT
        A time.Duration string that specifies the timeout for failing to
        establish a connection.

    X_CSI_SERIAL_VOL_ACCESS_ETCD_DIAL_KEEP_ALIVE_TIME
        A time.Duration string that defines the time after which the client
        pings the server to see if the transport is alive.

    X_CSI_SERIAL_VOL_ACCESS_ETCD_DIAL_KEEP_ALIVE_TIMEOUT
        A time.Duration string that defines the time that the client waits for
        a response for the keep-alive probe. If the response is not received
        in this time, the connection is closed.

    X_CSI_SERIAL_VOL_ACCESS_ETCD_MAX_CALL_SEND_MSG_SZ
        Defines the client-side request send limit in bytes. If 0, it defaults
        to 2.0 MiB (2 * 1024 * 1024). Make sure that "MaxCallSendMsgSize" <
        server-side default send/recv limit. ("--max-request-bytes" flag to
        etcd or "embed.Config.MaxRequestBytes").

    X_CSI_SERIAL_VOL_ACCESS_ETCD_MAX_CALL_RECV_MSG_SZ
        Defines the client-side response receive limit. If 0, it defaults to
        "math.MaxInt32", because range response can easily exceed request send
        limits. Make sure that "MaxCallRecvMsgSize" >= server-side default
        send/recv limit. ("--max-request-bytes" flag to etcd or
        "embed.Config.MaxRequestBytes").

    X_CSI_SERIAL_VOL_ACCESS_ETCD_USERNAME
        The user name used for authentication.

    X_CSI_SERIAL_VOL_ACCESS_ETCD_PASSWORD
        The password used for authentication.

    X_CSI_SERIAL_VOL_ACCESS_ETCD_REJECT_OLD_CLUSTER
        A flag that indicates refusal to create a client against an outdated
        cluster.

    X_CSI_SERIAL_VOL_ACCESS_ETCD_TLS
        A flag that indicates the client should attempt a TLS connection.

    X_CSI_SERIAL_VOL_ACCESS_ETCD_TLS_INSECURE
        A flag that indicates the TLS connection should not verify peer
        certificates.`
