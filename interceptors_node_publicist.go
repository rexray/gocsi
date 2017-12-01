package gocsi

import (
	"context"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/thecodeteam/gofsutil"
	xctx "golang.org/x/net/context"
)

// NodePublicistProvider is an interface that guards the NodePublishVolume RPC
// against invalid requests.
type NodePublicistProvider interface {
	// GetPublishedMountInfo should return the mount table entries
	// for a volume on a node host. If the volume is not mounted then
	// a nil or empty slice should be returned.
	GetPublishedMountInfo(
		ctx context.Context,
		id string, pubVolInfo map[string]string) ([]gofsutil.Info, error)
}

// NodePublicistOption configures the NodePublicist interceptor.
type NodePublicistOption func(*nodePublicistOpts)

type nodePublicistOpts struct {
	multiMount bool
}

// WithNodeMultiMount is a NodePublicistOption that allows a volume
// to be published on a single node host at different target paths.
func WithNodeMultiMount() NodePublicistOption {
	return func(o *nodePublicistOpts) {
		o.multiMount = true
	}
}

// NewNodePublicist returns a new server-side, gRPC interceptor
// that can be used in conjunction with a NodePublicistProvider to
// guard NodePublishVolume RPCs against invalid requests.
func NewNodePublicist(
	p NodePublicistProvider,
	opts ...NodePublicistOption) grpc.UnaryServerInterceptor {

	i := &nodePublicistInterceptor{p: p}

	// Configure the interceptor's options.
	for _, setOpt := range opts {
		setOpt(&i.opts)
	}

	return i.handle

}

type nodePublicistInterceptor struct {
	p    NodePublicistProvider
	opts nodePublicistOpts
}

func (i *nodePublicistInterceptor) handle(
	ctx xctx.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	treq, ok := req.(*csi.NodePublishVolumeRequest)
	if !ok {
		return handler(ctx, req)
	}

	if err := i.nodePublish(ctx, treq); err != nil {
		return nil, err
	}

	return handler(ctx, req)
}

func (i *nodePublicistInterceptor) nodePublish(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest) error {

	volCap := req.VolumeCapability
	if volCap == nil {
		return ErrVolumeCapabilityRequired
	}

	// If the request's Readonly field is true then the request's
	// volume capability be Block.
	reqRO := req.Readonly
	if _, ok := volCap.AccessType.(*csi.VolumeCapability_Block); ok && reqRO {
		return ErrReadOnlyInvalidForBlock
	}
	var (
		multiMountRO = i.opts.multiMount
		multiMountRW = i.opts.multiMount
	)

	if !multiMountRO || !multiMountRW {
		if volCap.AccessMode == nil {
			return ErrAccessModeRequired
		}
		switch req.VolumeCapability.AccessMode.Mode {
		case csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY:
			multiMountRO = true
		case csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER,
			csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER:
			multiMountRO = true
			multiMountRW = true
		}
	}

	mountInfo, err := i.p.GetPublishedMountInfo(
		ctx, req.VolumeId, req.PublishVolumeInfo)
	if err != nil {
		return err
	}

	if len(mountInfo) == 0 {
		return nil
	}

	// Make the target path absolute with no symlinks.
	reqTargetPath := req.TargetPath
	if !filepath.IsAbs(reqTargetPath) {
		v, err := filepath.Abs(reqTargetPath)
		if err != nil {
			return err
		}
		reqTargetPath = v
	}
	if err := gofsutil.EvalSymlinks(ctx, &reqTargetPath); err != nil {
		return err
	}

	// If there is a single mount entry for this volume check to see
	// if this is an idempotent operation.
	if len(mountInfo) == 1 {
		isPub, err := isPublishedOnNode(
			ctx, req.VolumeId, reqTargetPath, req.Readonly, mountInfo[0])
		if err != nil {
			return err
		}
		if isPub {
			return nil
		}
	}

	// Check to see if the request's Readonly flag is compatible with the
	// request's access mode.
	if (multiMountRW && !reqRO) || ((multiMountRO || multiMountRW) && reqRO) {
		return nil
	}

	return status.Errorf(codes.Aborted,
		"volume %s already published",
		req.VolumeId)
}

func isReadOnly(opts []string) (bool, string) {
	for _, o := range opts {
		if o == "ro" {
			return true, o
		}
		if o == "rw" {
			return false, o
		}
	}
	return false, "rw"
}

// isPublishedOnNode determines if the requested volume is already
// published on the node with the same rw option.
func isPublishedOnNode(
	ctx context.Context,
	volumeID, targetPath string, readOnly bool,
	info gofsutil.Info) (bool, error) {

	if info.Path != targetPath {
		return false, nil
	}

	ro, szRW := isReadOnly(info.Opts)

	if ro && readOnly {
		log.WithField("volumeId", volumeID).Debug(
			"idempotent node publish: ro")
		return true, nil
	}

	if !ro && !readOnly {
		log.WithField("volumeId", volumeID).Debug(
			"idempotent node publish: rw")
		return true, nil
	}

	return false, status.Errorf(codes.Aborted,
		"volume %s already published to %s as %s",
		volumeID, info.Path, szRW)
}
