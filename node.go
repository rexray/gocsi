package gocsi

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/codedellemc/gocsi/csi"
)

const (
	// FMGetNodeID is the full method name for the
	// eponymous RPC message.
	FMGetNodeID = "/" + Namespace +
		".Node/" +
		"GetNodeID"

	// FMNodePublishVolume is the full method name for the
	// eponymous RPC message.
	FMNodePublishVolume = "/" + Namespace +
		".Node/" +
		"NodePublishVolume"

	// FMNodeUnpublishVolume is the full method name for the
	// eponymous RPC message.
	FMNodeUnpublishVolume = "/" + Namespace +
		".Node/" +
		"NodeUnpublishVolume"

	// FMProbeNode is the full method name for the
	// eponymous RPC message.
	FMProbeNode = "/" + Namespace +
		".Node/" +
		"ProbeNode"

	// FMNodeGetCapabilities is the full method name for the
	// eponymous RPC message.
	FMNodeGetCapabilities = "/" + Namespace +
		".Node/" +
		"NodeGetCapabilities"
)

// GetNodeID issues a
// GetNodeID request
// to a CSI controller.
func GetNodeID(
	ctx context.Context,
	c csi.NodeClient,
	version *csi.Version,
	callOpts ...grpc.CallOption) (*csi.NodeID, error) {

	if version == nil {
		return nil, ErrVersionRequired
	}

	req := &csi.GetNodeIDRequest{
		Version: version,
	}

	res, err := c.GetNodeID(ctx, req, callOpts...)
	if err != nil {
		return nil, err
	}

	return res.GetResult().NodeId, nil
}

// NodePublishVolume issues a
// NodePublishVolume request
// to a CSI controller.
func NodePublishVolume(
	ctx context.Context,
	c csi.NodeClient,
	version *csi.Version,
	volumeID *csi.VolumeID,
	volumeMetadata *csi.VolumeMetadata,
	publishVolumeInfo *csi.PublishVolumeInfo,
	targetPath string,
	volumeCapability *csi.VolumeCapability,
	readonly bool,
	callOpts ...grpc.CallOption) error {

	if version == nil {
		return ErrVersionRequired
	}

	if volumeID == nil {
		return ErrVolumeIDRequired
	}

	if volumeCapability == nil {
		return ErrVolumeCapabilityRequired
	}

	if targetPath == "" {
		return ErrInvalidTargetPath
	}

	req := &csi.NodePublishVolumeRequest{
		Version:           version,
		VolumeId:          volumeID,
		VolumeMetadata:    volumeMetadata,
		PublishVolumeInfo: publishVolumeInfo,
		TargetPath:        targetPath,
		VolumeCapability:  volumeCapability,
		Readonly:          readonly,
	}

	_, err := c.NodePublishVolume(ctx, req, callOpts...)
	if err != nil {
		return err
	}

	return nil
}

// NodeUnpublishVolume issues a
// NodeUnpublishVolume request
// to a CSI controller.
func NodeUnpublishVolume(
	ctx context.Context,
	c csi.NodeClient,
	version *csi.Version,
	volumeID *csi.VolumeID,
	volumeMetadata *csi.VolumeMetadata,
	targetPath string,
	callOpts ...grpc.CallOption) error {

	if version == nil {
		return ErrVersionRequired
	}

	if volumeID == nil {
		return ErrVolumeIDRequired
	}

	if targetPath == "" {
		return ErrInvalidTargetPath
	}

	req := &csi.NodeUnpublishVolumeRequest{
		Version:        version,
		VolumeId:       volumeID,
		VolumeMetadata: volumeMetadata,
		TargetPath:     targetPath,
	}

	_, err := c.NodeUnpublishVolume(ctx, req, callOpts...)
	if err != nil {
		return err
	}

	return nil
}

// ProbeNode issues a
// ProbeNode request
// to a CSI controller.
func ProbeNode(
	ctx context.Context,
	c csi.NodeClient,
	version *csi.Version,
	callOpts ...grpc.CallOption) error {

	if version == nil {
		return ErrVersionRequired
	}

	req := &csi.ProbeNodeRequest{
		Version: version,
	}

	_, err := c.ProbeNode(ctx, req, callOpts...)
	if err != nil {
		return err
	}

	return nil
}

// NodeGetCapabilities issues a NodeGetCapabilities request to a
// CSI controller.
func NodeGetCapabilities(
	ctx context.Context,
	c csi.NodeClient,
	version *csi.Version,
	callOpts ...grpc.CallOption) (
	capabilties []*csi.NodeServiceCapability, err error) {

	if version == nil {
		return nil, ErrVersionRequired
	}

	req := &csi.NodeGetCapabilitiesRequest{
		Version: version,
	}

	res, err := c.NodeGetCapabilities(ctx, req, callOpts...)
	if err != nil {
		return nil, err
	}

	return res.GetResult().Capabilities, nil
}
