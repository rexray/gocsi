package gocsi

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/thecodeteam/gocsi/csi"
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

	// FMNodeProbe is the full method name for the
	// eponymous RPC message.
	FMNodeProbe = "/" + Namespace +
		".Node/" +
		"NodeProbe"

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
	callOpts ...grpc.CallOption) (string, error) {

	req := &csi.GetNodeIDRequest{
		Version: version,
	}

	res, err := c.GetNodeID(ctx, req, callOpts...)
	if err != nil {
		return "", err
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
	volumeID string,
	volumeAttribs map[string]string,
	publishVolumeInfo map[string]string,
	targetPath string,
	volumeCapability *csi.VolumeCapability,
	readonly bool,
	callOpts ...grpc.CallOption) error {

	req := &csi.NodePublishVolumeRequest{
		Version:           version,
		VolumeId:          volumeID,
		VolumeAttributes:  volumeAttribs,
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
	volumeID string,
	targetPath string,
	callOpts ...grpc.CallOption) error {

	req := &csi.NodeUnpublishVolumeRequest{
		Version:    version,
		VolumeId:   volumeID,
		TargetPath: targetPath,
	}

	_, err := c.NodeUnpublishVolume(ctx, req, callOpts...)
	if err != nil {
		return err
	}

	return nil
}

// NodeProbe issues a
// NodeProbe request
// to a CSI controller.
func NodeProbe(
	ctx context.Context,
	c csi.NodeClient,
	version *csi.Version,
	callOpts ...grpc.CallOption) error {

	req := &csi.NodeProbeRequest{
		Version: version,
	}

	_, err := c.NodeProbe(ctx, req, callOpts...)
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

	req := &csi.NodeGetCapabilitiesRequest{
		Version: version,
	}

	res, err := c.NodeGetCapabilities(ctx, req, callOpts...)
	if err != nil {
		return nil, err
	}

	return res.GetResult().Capabilities, nil
}
