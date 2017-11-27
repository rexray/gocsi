package gocsi

import (
	"context"
	"sync"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	xctx "golang.org/x/net/context"
)

// IdempotencyProvider is the interface that works with a server-side,
// gRPC interceptor to provide serial access and idempotency for CSI's
// volume resources.
type IdempotencyProvider interface {
	// GetVolumeID should return the ID of the volume specified
	// by the provided volume name. If the volume does not exist then
	// an empty string should be returned.
	GetVolumeID(ctx context.Context, name string) (string, error)

	// GetVolumeInfo should return information about the volume
	// specified by the provided volume ID or name. If the volume does not
	// exist then a nil value should be returned.
	GetVolumeInfo(ctx context.Context, id, name string) (*csi.VolumeInfo, error)

	// IsControllerPublished should return publication for a volume's
	// publication status on a specified node.
	IsControllerPublished(
		ctx context.Context,
		volumeID, nodeID string) (map[string]string, error)

	// IsNodePublished should return a flag indicating whether or
	// not the volume exists and is published on the current host.
	IsNodePublished(
		ctx context.Context,
		id string,
		pubVolInfo map[string]string,
		targetPath string) (bool, error)
}

// IdempotentInterceptorOption configures the idempotent interceptor.
type IdempotentInterceptorOption func(*idempIntercOpts)

type idempIntercOpts struct {
	timeout       time.Duration
	requireVolume bool
}

// WithIdempTimeout is an IdempotentInterceptorOption that sets the
// timeout used by the idempotent interceptor.
func WithIdempTimeout(t time.Duration) IdempotentInterceptorOption {
	return func(o *idempIntercOpts) {
		o.timeout = t
	}
}

// WithIdempRequireVolumeExists is an IdempotentInterceptorOption that
// enforces the requirement that volumes must exist before proceeding
// with an operation.
func WithIdempRequireVolumeExists() IdempotentInterceptorOption {
	return func(o *idempIntercOpts) {
		o.requireVolume = true
	}
}

// NewIdempotentInterceptor returns a new server-side, gRPC interceptor
// that can be used in conjunction with an IdempotencyProvider to
// provide serialized, idempotent access to the following CSI RPCs:
//
//  * CreateVolume
//  * DeleteVolume
//  * ControllerPublishVolume
//  * ControllerUnpublishVolume
//  * NodePublishVolume
//  * NodeUnpublishVolume
func NewIdempotentInterceptor(
	p IdempotencyProvider,
	opts ...IdempotentInterceptorOption) grpc.UnaryServerInterceptor {

	i := &idempotencyInterceptor{
		p:            p,
		volIDLocks:   map[string]*volLockInfo{},
		volNameLocks: map[string]*volLockInfo{},
	}

	// Configure the idempotent interceptor's options.
	for _, setOpt := range opts {
		setOpt(&i.opts)
	}

	return i.handle
}

type volLockInfo struct {
	MutexWithTryLock
	methodInErr map[string]struct{}
}

type idempotencyInterceptor struct {
	p             IdempotencyProvider
	volIDLocksL   sync.Mutex
	volNameLocksL sync.Mutex
	volIDLocks    map[string]*volLockInfo
	volNameLocks  map[string]*volLockInfo
	opts          idempIntercOpts
}

func (i *idempotencyInterceptor) lockWithID(id string) *volLockInfo {
	i.volIDLocksL.Lock()
	defer i.volIDLocksL.Unlock()
	lock := i.volIDLocks[id]
	if lock == nil {
		lock = &volLockInfo{
			MutexWithTryLock: NewMutexWithTryLock(),
			methodInErr:      map[string]struct{}{},
		}
		i.volIDLocks[id] = lock
	}
	return lock
}

func (i *idempotencyInterceptor) lockWithName(name string) *volLockInfo {
	i.volNameLocksL.Lock()
	defer i.volNameLocksL.Unlock()
	lock := i.volNameLocks[name]
	if lock == nil {
		lock = &volLockInfo{
			MutexWithTryLock: NewMutexWithTryLock(),
			methodInErr:      map[string]struct{}{},
		}
		i.volNameLocks[name] = lock
	}
	return lock
}

func isOpPending(err error) bool {
	stat, ok := status.FromError(err)
	return ok &&
		stat.Code() == codes.FailedPrecondition &&
		stat.Message() == "op pending"
}

func (i *idempotencyInterceptor) handle(
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

func (i *idempotencyInterceptor) controllerPublishVolume(
	ctx context.Context,
	req *csi.ControllerPublishVolumeRequest,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (res interface{}, resErr error) {

	lock := i.lockWithID(req.VolumeId)
	if !lock.TryLock(i.opts.timeout) {
		return nil, ErrOpPending
	}

	// At the end of this function check for a response error or if
	// the response itself contains an error. If either is true then
	// mark the current method as in error.
	//
	// If neither is true then check to see if the method has been
	// marked in error in the past and remove that mark to reclaim
	// memory.
	defer func() {
		if resErr != nil {
			lock.methodInErr[info.FullMethod] = struct{}{}
		} else if _, ok := lock.methodInErr[info.FullMethod]; ok {
			delete(lock.methodInErr, info.FullMethod)
		}
	}()
	defer lock.Unlock()

	// If the method has been marked in error then it means a previous
	// call to this function returned an error. In these cases a
	// subsequent call should bypass idempotency.
	if _, ok := lock.methodInErr[info.FullMethod]; ok {
		return handler(ctx, req)
	}

	// If configured to do so, check to see if the volume exists and
	// return an error if it does not.
	if i.opts.requireVolume {
		volInfo, err := i.p.GetVolumeInfo(ctx, req.VolumeId, "")
		if err != nil {
			return nil, err
		}
		if volInfo == nil {
			return nil, status.Error(codes.NotFound, req.VolumeId)
		}
	}

	pubInfo, err := i.p.IsControllerPublished(ctx, req.VolumeId, req.NodeId)
	if err != nil {
		return nil, err
	}
	if pubInfo != nil {
		log.WithField("volumeID", req.VolumeId).Info(
			"idempotent controller publish")
		return &csi.ControllerPublishVolumeResponse{
			PublishVolumeInfo: pubInfo,
		}, nil
	}

	return handler(ctx, req)
}

func (i *idempotencyInterceptor) controllerUnpublishVolume(
	ctx context.Context,
	req *csi.ControllerUnpublishVolumeRequest,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (res interface{}, resErr error) {

	lock := i.lockWithID(req.VolumeId)
	if !lock.TryLock(i.opts.timeout) {
		return nil, ErrOpPending
	}

	// At the end of this function check for a response error or if
	// the response itself contains an error. If either is true then
	// mark the current method as in error.
	//
	// If neither is true then check to see if the method has been
	// marked in error in the past and remove that mark to reclaim
	// memory.
	defer func() {
		if resErr != nil {
			lock.methodInErr[info.FullMethod] = struct{}{}
		} else if _, ok := lock.methodInErr[info.FullMethod]; ok {
			delete(lock.methodInErr, info.FullMethod)
		}
	}()
	defer lock.Unlock()

	// If the method has been marked in error then it means a previous
	// call to this function returned an error. In these cases a
	// subsequent call should bypass idempotency.
	if _, ok := lock.methodInErr[info.FullMethod]; ok {
		return handler(ctx, req)
	}

	// If configured to do so, check to see if the volume exists and
	// return an error if it does not.
	if i.opts.requireVolume {
		volInfo, err := i.p.GetVolumeInfo(ctx, req.VolumeId, "")
		if err != nil {
			return nil, err
		}
		if volInfo == nil {
			return nil, status.Error(codes.NotFound, req.VolumeId)
		}
	}

	pubInfo, err := i.p.IsControllerPublished(ctx, req.VolumeId, req.NodeId)
	if err != nil {
		return nil, err
	}
	if pubInfo == nil {
		log.WithField("volumeID", req.VolumeId).Info(
			"idempotent controller unpublish")
		return &csi.ControllerUnpublishVolumeResponse{}, nil
	}

	return handler(ctx, req)
}

func (i *idempotencyInterceptor) createVolume(
	ctx context.Context,
	req *csi.CreateVolumeRequest,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (res interface{}, resErr error) {

	reqID, _ := GetRequestID(ctx)
	fields := map[string]interface{}{
		"requestID":  reqID,
		"volumeName": req.Name,
	}

	log.WithFields(fields).Debug("idemp: begin createVolume")

	// First attempt to lock the volume by the provided name. If no lock
	// can be obtained then exit with the appropriate error.
	nameLock := i.lockWithName(req.Name)
	if !nameLock.TryLock(i.opts.timeout) {
		return nil, ErrOpPending
	}

	// At the end of this function check for a response error or if
	// the response itself contains an error. If either is true then
	// mark the current method as in error.
	//
	// If neither is true then check to see if the method has been
	// marked in error in the past and remove that mark to reclaim
	// memory.
	defer func() {
		if resErr != nil {

			// Check to see if the error code indicates an operation is
			// pending for this resource. If it is then do not mark this
			// method in error.
			if isOpPending(resErr) {
				return
			}
			nameLock.methodInErr[info.FullMethod] = struct{}{}
		} else if _, ok := nameLock.methodInErr[info.FullMethod]; ok {
			delete(nameLock.methodInErr, info.FullMethod)
		}
	}()
	defer nameLock.Unlock()

	// If the method has been marked in error then it means a previous
	// call to this function returned an error. In these cases a
	// subsequent call should bypass idempotency.
	if _, ok := nameLock.methodInErr[info.FullMethod]; ok {
		log.WithFields(fields).Debug("creating volume: nameInErr")
		return handler(ctx, req)
	}

	// Next, attempt to get the volume info based on the name.
	volInfo, err := i.p.GetVolumeInfo(ctx, "", req.Name)
	if err != nil {
		return nil, err
	}

	// If the volInfo is nil then it means the volume does not exist.
	// Return early, passing control to the next handler in the chain.
	if volInfo == nil {
		log.WithFields(fields).Debug("creating volume")
		return handler(ctx, req)
	}

	// If the volInfo is not nil it means the volume already exists.
	// The volume info contains the volume's ID. Use that to obtain a
	// volume ID-based lock for the volume.
	idLock := i.lockWithID(volInfo.Id)
	if !idLock.TryLock(i.opts.timeout) {
		return nil, ErrOpPending
	}

	// At the end of this function check for a response error or if
	// the response itself contains an error. If either is true then
	// mark the current method as in error.
	//
	// If neither is true then check to see if the method has been
	// marked in error in the past and remove that mark to reclaim
	// memory.
	defer func() {
		if resErr != nil {
			idLock.methodInErr[info.FullMethod] = struct{}{}
		} else if _, ok := idLock.methodInErr[info.FullMethod]; ok {
			delete(idLock.methodInErr, info.FullMethod)
		}
	}()
	defer idLock.Unlock()

	// If the method has been marked in error then it means a previous
	// call to this function returned an error. In these cases a
	// subsequent call should bypass idempotency.
	if _, ok := idLock.methodInErr[info.FullMethod]; ok {
		log.WithFields(fields).Debug("creating volume: idInErr")
		return handler(ctx, req)
	}

	// The ID lock has been obtained. Once again call GetVolumeInfo,
	// this time with the volume ID, now that the ID lock is held.
	// This ensures the volume still exists since it could have been
	// removed in the time it took to obtain the ID lock.
	volInfo, err = i.p.GetVolumeInfo(ctx, volInfo.Id, "")
	if err != nil {
		return nil, err
	}

	// If the volume info is nil it means the volume was removed in
	// the time it took to obtain the lock ID. Return early, passing
	// control to the next handler in the chain.
	if volInfo == nil {
		log.WithFields(fields).Debug("creating volume: 2nd attempt")
		return handler(ctx, req)
	}

	// If the volume info still exists then it means the volume
	// exists! Go ahead and return the volume info and note this
	// as an idempotent create call.
	log.WithFields(map[string]interface{}{
		"requestID":  reqID,
		"volumeID":   volInfo.Id,
		"volumeName": req.Name}).Info("idempotent create")
	return &csi.CreateVolumeResponse{
		VolumeInfo: volInfo,
	}, nil
}

func (i *idempotencyInterceptor) deleteVolume(
	ctx context.Context,
	req *csi.DeleteVolumeRequest,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (res interface{}, resErr error) {

	lock := i.lockWithID(req.VolumeId)
	if !lock.TryLock(i.opts.timeout) {
		return nil, ErrOpPending
	}

	// At the end of this function check for a response error or if
	// the response itself contains an error. If either is true then
	// mark the current method as in error.
	//
	// If neither is true then check to see if the method has been
	// marked in error in the past and remove that mark to reclaim
	// memory.
	defer func() {
		if resErr != nil {
			lock.methodInErr[info.FullMethod] = struct{}{}
		} else if _, ok := lock.methodInErr[info.FullMethod]; ok {
			delete(lock.methodInErr, info.FullMethod)
		}
	}()
	defer lock.Unlock()

	// If the method has been marked in error then it means a previous
	// call to this function returned an error. In these cases a
	// subsequent call should bypass idempotency.
	if _, ok := lock.methodInErr[info.FullMethod]; ok {
		return handler(ctx, req)
	}

	// If configured to do so, check to see if the volume exists and
	// return an error if it does not.
	var volExists bool
	if i.opts.requireVolume {
		volInfo, err := i.p.GetVolumeInfo(ctx, req.VolumeId, "")
		if err != nil {
			return nil, err
		}
		if volInfo == nil {
			log.WithField("volumeID", req.VolumeId).Info("idempotent delete.a")
			return nil, status.Error(codes.NotFound, req.VolumeId)
		}
		volExists = true
	}

	// Check to see if the volume exists if that has not yet been
	// verified above.
	if !volExists {
		volInfo, err := i.p.GetVolumeInfo(ctx, req.VolumeId, "")
		if err != nil {
			return nil, err
		}
		volExists = volInfo != nil
	}

	// Indicate an idempotent delete operation if the volume does not exist.
	if !volExists {
		log.WithField("volumeID", req.VolumeId).Info("idempotent delete.b")
		return nil, status.Error(codes.NotFound, req.VolumeId)
	}

	return handler(ctx, req)
}

func (i *idempotencyInterceptor) nodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (res interface{}, resErr error) {

	lock := i.lockWithID(req.VolumeId)
	if !lock.TryLock(i.opts.timeout) {
		return nil, ErrOpPending
	}

	// At the end of this function check for a response error or if
	// the response itself contains an error. If either is true then
	// mark the current method as in error.
	//
	// If neither is true then check to see if the method has been
	// marked in error in the past and remove that mark to reclaim
	// memory.
	defer func() {
		if resErr != nil {
			lock.methodInErr[info.FullMethod] = struct{}{}
		} else if _, ok := lock.methodInErr[info.FullMethod]; ok {
			delete(lock.methodInErr, info.FullMethod)
		}
	}()
	defer lock.Unlock()

	// If the method has been marked in error then it means a previous
	// call to this function returned an error. In these cases a
	// subsequent call should bypass idempotency.
	if _, ok := lock.methodInErr[info.FullMethod]; ok {
		return handler(ctx, req)
	}

	// If configured to do so, check to see if the volume exists and
	// return an error if it does not.
	if i.opts.requireVolume {
		volInfo, err := i.p.GetVolumeInfo(ctx, req.VolumeId, "")
		if err != nil {
			return nil, err
		}
		if volInfo == nil {
			return nil, status.Error(codes.NotFound, req.VolumeId)
		}
	}

	ok, err := i.p.IsNodePublished(
		ctx, req.VolumeId, req.PublishVolumeInfo, req.TargetPath)
	if err != nil {
		return nil, err
	}
	if ok {
		log.WithField("volumeId", req.VolumeId).Info("idempotent node publish")
		return &csi.NodePublishVolumeResponse{}, nil
	}

	return handler(ctx, req)
}

func (i *idempotencyInterceptor) nodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (res interface{}, resErr error) {

	lock := i.lockWithID(req.VolumeId)
	if !lock.TryLock(i.opts.timeout) {
		return nil, ErrOpPending
	}

	// At the end of this function check for a response error or if
	// the response itself contains an error. If either is true then
	// mark the current method as in error.
	//
	// If neither is true then check to see if the method has been
	// marked in error in the past and remove that mark to reclaim
	// memory.
	defer func() {
		if resErr != nil {
			lock.methodInErr[info.FullMethod] = struct{}{}
		} else if _, ok := lock.methodInErr[info.FullMethod]; ok {
			delete(lock.methodInErr, info.FullMethod)
		}
	}()
	defer lock.Unlock()

	// If the method has been marked in error then it means a previous
	// call to this function returned an error. In these cases a
	// subsequent call should bypass idempotency.
	if _, ok := lock.methodInErr[info.FullMethod]; ok {
		return handler(ctx, req)
	}

	// If configured to do so, check to see if the volume exists and
	// return an error if it does not.
	if i.opts.requireVolume {
		volInfo, err := i.p.GetVolumeInfo(ctx, req.VolumeId, "")
		if err != nil {
			return nil, err
		}
		if volInfo == nil {
			return nil, status.Error(codes.NotFound, req.VolumeId)
		}
	}

	ok, err := i.p.IsNodePublished(ctx, req.VolumeId, nil, req.TargetPath)
	if err != nil {
		return nil, err
	}
	if !ok {
		log.WithField("volumeId", req.VolumeId).Info(
			"idempotent node unpublish")
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}

	return handler(ctx, req)
}
