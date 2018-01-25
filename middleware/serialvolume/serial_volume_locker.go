package serialvolume

import (
	"context"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/thecodeteam/gosync"
	xctx "golang.org/x/net/context"
	"google.golang.org/grpc"

	csierr "github.com/thecodeteam/gocsi/errors"
)

// VolumeLockerProvider is able to provide gosync.TryLocker objects for
// volumes by ID and name.
type VolumeLockerProvider interface {
	// GetLockWithID gets a lock for a volume with provided ID. If a lock
	// for the specified volume ID does not exist then a new lock is created
	// and returned.
	GetLockWithID(ctx context.Context, id string) (gosync.TryLocker, error)

	// GetLockWithName gets a lock for a volume with provided name. If a lock
	// for the specified volume name does not exist then a new lock is created
	// and returned.
	GetLockWithName(ctx context.Context, name string) (gosync.TryLocker, error)
}

// Option configures the interceptor.
type Option func(*opts)

type opts struct {
	timeout time.Duration
	locker  VolumeLockerProvider
}

// WithTimeout is an Option that sets the timeout used by the interceptor.
func WithTimeout(t time.Duration) Option {
	return func(o *opts) {
		o.timeout = t
	}
}

// New returns a new server-side, gRPC interceptor
// that provides serial access to volume resources across the following
// RPCs:
//
//  * CreateVolume
//  * DeleteVolume
//  * ControllerPublishVolume
//  * ControllerUnpublishVolume
//  * NodePublishVolume
//  * NodeUnpublishVolume
func New(opts ...Option) grpc.UnaryServerInterceptor {

	i := &interceptor{}

	// Configure the interceptor's options.
	for _, setOpt := range opts {
		setOpt(&i.opts)
	}

	// If no lock provider is configured then set the default,
	// in-memory provider.
	if i.opts.locker == nil {
		i.opts.locker = &defaultLockProvider{
			volIDLocks:   map[string]gosync.TryLocker{},
			volNameLocks: map[string]gosync.TryLocker{},
		}
	}

	return i.handle
}

type interceptor struct {
	opts opts
}

func (i *interceptor) handle(
	ctx xctx.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	switch treq := req.(type) {
	case *csi.ControllerPublishVolumeRequest:
		return i.controllerPublishVolume(ctx, treq, info, handler)
	case *csi.ControllerUnpublishVolumeRequest:
		return i.controllerUnpublishVolume(ctx, treq, info, handler)
	case *csi.CreateVolumeRequest:
		return i.createVolume(ctx, treq, info, handler)
	case *csi.DeleteVolumeRequest:
		return i.deleteVolume(ctx, treq, info, handler)
	case *csi.NodePublishVolumeRequest:
		return i.nodePublishVolume(ctx, treq, info, handler)
	case *csi.NodeUnpublishVolumeRequest:
		return i.nodeUnpublishVolume(ctx, treq, info, handler)
	}

	return handler(ctx, req)
}

func (i *interceptor) controllerPublishVolume(
	ctx context.Context,
	req *csi.ControllerPublishVolumeRequest,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (res interface{}, resErr error) {

	lock, err := i.opts.locker.GetLockWithID(ctx, req.VolumeId)
	if err != nil {
		return nil, err
	}
	if !lock.TryLock(i.opts.timeout) {
		return nil, csierr.ErrOpPending
	}
	defer lock.Unlock()

	return handler(ctx, req)
}

func (i *interceptor) controllerUnpublishVolume(
	ctx context.Context,
	req *csi.ControllerUnpublishVolumeRequest,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (res interface{}, resErr error) {

	lock, err := i.opts.locker.GetLockWithID(ctx, req.VolumeId)
	if err != nil {
		return nil, err
	}
	if !lock.TryLock(i.opts.timeout) {
		return nil, csierr.ErrOpPending
	}
	defer lock.Unlock()

	return handler(ctx, req)
}

func (i *interceptor) createVolume(
	ctx context.Context,
	req *csi.CreateVolumeRequest,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (res interface{}, resErr error) {

	lock, err := i.opts.locker.GetLockWithName(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if !lock.TryLock(i.opts.timeout) {
		return nil, csierr.ErrOpPending
	}
	defer lock.Unlock()

	return handler(ctx, req)
}

func (i *interceptor) deleteVolume(
	ctx context.Context,
	req *csi.DeleteVolumeRequest,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (res interface{}, resErr error) {

	lock, err := i.opts.locker.GetLockWithID(ctx, req.VolumeId)
	if err != nil {
		return nil, err
	}
	if !lock.TryLock(i.opts.timeout) {
		return nil, csierr.ErrOpPending
	}
	defer lock.Unlock()

	return handler(ctx, req)
}

func (i *interceptor) nodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (res interface{}, resErr error) {

	lock, err := i.opts.locker.GetLockWithID(ctx, req.VolumeId)
	if err != nil {
		return nil, err
	}
	if !lock.TryLock(i.opts.timeout) {
		return nil, csierr.ErrOpPending
	}
	defer lock.Unlock()

	return handler(ctx, req)
}

func (i *interceptor) nodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (res interface{}, resErr error) {

	lock, err := i.opts.locker.GetLockWithID(ctx, req.VolumeId)
	if err != nil {
		return nil, err
	}
	if !lock.TryLock(i.opts.timeout) {
		return nil, csierr.ErrOpPending
	}
	defer lock.Unlock()

	return handler(ctx, req)
}
