package serialvolume

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/akutz/gosync"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"

	csienv "github.com/rexray/gocsi/env"
)

const pending = "pending"

// LockProvider is able to provide gosync.TryLocker objects for
// volumes by ID and name.
type LockProvider interface {
	// GetLockWithID gets a lock for a volume with provided ID. If a lock
	// for the specified volume ID does not exist then a new lock is created
	// and returned.
	GetLockWithID(ctx context.Context, id string) (gosync.TryLocker, error)

	// GetLockWithName gets a lock for a volume with provided name. If a lock
	// for the specified volume name does not exist then a new lock is created
	// and returned.
	GetLockWithName(ctx context.Context, name string) (gosync.TryLocker, error)
}

type hasUsage interface {
	// Usage returns the lock provider's usage string.
	Usage() string
}

// Middleware provides serial volume access.
type Middleware struct {
	sync.Once
	Timeout      time.Duration
	LockProvider LockProvider
}

// Init is available to explicitly initialize the middleware.
func (i *Middleware) Init(ctx context.Context) (err error) {
	return i.initOnce(ctx)
}

func (i *Middleware) initOnce(ctx context.Context) (err error) {
	i.Once.Do(func() {
		err = i.init(ctx)
	})
	return
}

func (i *Middleware) init(ctx context.Context) error {
	if v, ok := csienv.LookupEnv(ctx, "X_CSI_SERIAL_VOL_ACCESS_TIMEOUT"); ok {
		i.Timeout, _ = time.ParseDuration(v)
	}

	log.WithFields(map[string]interface{}{
		"Timeout":      i.Timeout,
		"LockProvider": fmt.Sprintf("%T", i.LockProvider),
	}).Info("middleware: serial volume access")

	return nil
}

// ErrNoLockProvider occurs when no lock provider is assigned to the
// SerialVolumeAccess middleware.
var ErrNoLockProvider = errors.New("no volume lock provider")

// HandleServer is a server-side, gRPC interceptor that provides serial
// access to volume resources.
func (i *Middleware) HandleServer(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	if i.LockProvider == nil {
		return nil, ErrNoLockProvider
	}

	switch treq := req.(type) {
	case *csi.ControllerPublishVolumeRequest:
		return i.controllerPublishVolume(ctx, treq, info, handler)
	case *csi.ControllerUnpublishVolumeRequest:
		return i.controllerUnpublishVolume(ctx, treq, info, handler)
	case *csi.CreateVolumeRequest:
		return i.createVolume(ctx, treq, info, handler)
	case *csi.DeleteVolumeRequest:
		return i.deleteVolume(ctx, treq, info, handler)
	case *csi.NodeStageVolumeRequest:
		return i.nodeStageVolume(ctx, treq, info, handler)
	case *csi.NodeUnstageVolumeRequest:
		return i.nodeUnstageVolume(ctx, treq, info, handler)
	case *csi.NodePublishVolumeRequest:
		return i.nodePublishVolume(ctx, treq, info, handler)
	case *csi.NodeUnpublishVolumeRequest:
		return i.nodeUnpublishVolume(ctx, treq, info, handler)
	}

	return handler(ctx, req)
}

func (i *Middleware) controllerPublishVolume(
	ctx context.Context,
	req *csi.ControllerPublishVolumeRequest,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (res interface{}, resErr error) {

	lock, err := i.LockProvider.GetLockWithID(ctx, req.VolumeId)
	if err != nil {
		return nil, err
	}
	if closer, ok := lock.(io.Closer); ok {
		defer closer.Close()
	}
	if !lock.TryLock(i.Timeout) {
		return nil, status.Error(codes.Aborted, pending)
	}
	defer lock.Unlock()

	return handler(ctx, req)
}

func (i *Middleware) controllerUnpublishVolume(
	ctx context.Context,
	req *csi.ControllerUnpublishVolumeRequest,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (res interface{}, resErr error) {

	lock, err := i.LockProvider.GetLockWithID(ctx, req.VolumeId)
	if err != nil {
		return nil, err
	}
	if closer, ok := lock.(io.Closer); ok {
		defer closer.Close()
	}
	if !lock.TryLock(i.Timeout) {
		return nil, status.Error(codes.Aborted, pending)
	}
	defer lock.Unlock()

	return handler(ctx, req)
}

func (i *Middleware) createVolume(
	ctx context.Context,
	req *csi.CreateVolumeRequest,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (res interface{}, resErr error) {

	lock, err := i.LockProvider.GetLockWithName(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if closer, ok := lock.(io.Closer); ok {
		defer closer.Close()
	}
	if !lock.TryLock(i.Timeout) {
		return nil, status.Error(codes.Aborted, pending)
	}
	defer lock.Unlock()

	return handler(ctx, req)
}

func (i *Middleware) deleteVolume(
	ctx context.Context,
	req *csi.DeleteVolumeRequest,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (res interface{}, resErr error) {

	lock, err := i.LockProvider.GetLockWithID(ctx, req.VolumeId)
	if err != nil {
		return nil, err
	}
	if closer, ok := lock.(io.Closer); ok {
		defer closer.Close()
	}
	if !lock.TryLock(i.Timeout) {
		return nil, status.Error(codes.Aborted, pending)
	}
	defer lock.Unlock()

	return handler(ctx, req)
}

func (i *Middleware) nodeStageVolume(
	ctx context.Context,
	req *csi.NodeStageVolumeRequest,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (res interface{}, resErr error) {

	lock, err := i.LockProvider.GetLockWithID(ctx, req.VolumeId)
	if err != nil {
		return nil, err
	}
	if closer, ok := lock.(io.Closer); ok {
		defer closer.Close()
	}
	if !lock.TryLock(i.Timeout) {
		return nil, status.Error(codes.Aborted, pending)
	}
	defer lock.Unlock()

	return handler(ctx, req)
}

func (i *Middleware) nodeUnstageVolume(
	ctx context.Context,
	req *csi.NodeUnstageVolumeRequest,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (res interface{}, resErr error) {

	lock, err := i.LockProvider.GetLockWithID(ctx, req.VolumeId)
	if err != nil {
		return nil, err
	}
	if closer, ok := lock.(io.Closer); ok {
		defer closer.Close()
	}
	if !lock.TryLock(i.Timeout) {
		return nil, status.Error(codes.Aborted, pending)
	}
	defer lock.Unlock()

	return handler(ctx, req)
}

func (i *Middleware) nodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (res interface{}, resErr error) {

	lock, err := i.LockProvider.GetLockWithID(ctx, req.VolumeId)
	if err != nil {
		return nil, err
	}
	if closer, ok := lock.(io.Closer); ok {
		defer closer.Close()
	}
	if !lock.TryLock(i.Timeout) {
		return nil, status.Error(codes.Aborted, pending)
	}
	defer lock.Unlock()

	return handler(ctx, req)
}

func (i *Middleware) nodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (res interface{}, resErr error) {

	lock, err := i.LockProvider.GetLockWithID(ctx, req.VolumeId)
	if err != nil {
		return nil, err
	}
	if closer, ok := lock.(io.Closer); ok {
		defer closer.Close()
	}
	if !lock.TryLock(i.Timeout) {
		return nil, status.Error(codes.Aborted, pending)
	}
	defer lock.Unlock()

	return handler(ctx, req)
}

// Usage returns the middleware's usage string.
func (i *Middleware) Usage() string {
	if lp, ok := i.LockProvider.(hasUsage); ok {
		return fmt.Sprintf("%s\n\n%s", usage, lp.Usage())
	}
	return usage
}

const usage = `SERIAL VOLUME ACCESS
    X_CSI_SERIAL_VOL_ACCESS_TIMEOUT
        A time.Duration string that determines how long the serial volume
        access middleware waits to obtain a lock for the request's volume before
        returning a the gRPC error code FailedPrecondition (5) to indicate
        an operation is already pending for the specified volume.`
