package gocsi

import (
	"errors"
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/codedellemc/gocsi/csi"
)

const (
	fmNodePfx = "/" + Namespace + ".Node/"

	// FMNodePublishVolume is the eponymous, full method name.
	FMNodePublishVolume = fmNodePfx +
		"NodePublishVolume"
	// FMNodeUnpublishVolume is the eponymous, full method name.
	FMNodeUnpublishVolume = fmNodePfx +
		"NodeUnpublishVolume"
	// FMGetNodeID is the eponymous, full method name.
	FMGetNodeID = fmNodePfx +
		"GetNodeID"
	// FMProbeNode is the eponymous, full method name.
	FMProbeNode = fmNodePfx +
		"ProbeNode"
	// FMNodeGetCapabilities is the eponymous, full method name.
	FMNodeGetCapabilities = fmNodePfx +
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

	// check to see if there is a csi error
	if cerr := res.GetError(); cerr != nil {
		if err := cerr.GetGetNodeIdError(); err != nil {
			return nil, fmt.Errorf(
				"error: GetNodeID failed: %d: %s",
				err.GetErrorCode(),
				err.GetErrorDescription())
		}
		if err := cerr.GetGeneralError(); err != nil {
			return nil, fmt.Errorf(
				"error: GetNodeID failed: %d: %s",
				err.GetErrorCode(),
				err.GetErrorDescription())
		}
		return nil, errors.New(cerr.String())
	}

	result := res.GetResult()
	if result == nil {
		return nil, ErrNilResult
	}

	data := result.GetNodeId()
	if data == nil {
		return nil, ErrNilNodeID
	}

	return data, nil
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

	res, err := c.NodePublishVolume(ctx, req, callOpts...)
	if err != nil {
		return err
	}

	// check to see if there is a csi error
	if cerr := res.GetError(); cerr != nil {
		if err := cerr.GetNodePublishVolumeError(); err != nil {
			return fmt.Errorf(
				"error: NodePublishVolume failed: %d: %s",
				err.GetErrorCode(),
				err.GetErrorDescription())
		}
		if err := cerr.GetGeneralError(); err != nil {
			return fmt.Errorf(
				"error: NodePublishVolume failed: %d: %s",
				err.GetErrorCode(),
				err.GetErrorDescription())
		}
		return errors.New(cerr.String())
	}

	result := res.GetResult()
	if result == nil {
		return ErrNilResult
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

	res, err := c.NodeUnpublishVolume(ctx, req, callOpts...)
	if err != nil {
		return err
	}

	// check to see if there is a csi error
	if cerr := res.GetError(); cerr != nil {
		if err := cerr.GetNodeUnpublishVolumeError(); err != nil {
			return fmt.Errorf(
				"error: NodeUnpublishVolume failed: %d: %s",
				err.GetErrorCode(),
				err.GetErrorDescription())
		}
		if err := cerr.GetGeneralError(); err != nil {
			return fmt.Errorf(
				"error: NodeUnpublishVolume failed: %d: %s",
				err.GetErrorCode(),
				err.GetErrorDescription())
		}
		return errors.New(cerr.String())
	}

	result := res.GetResult()
	if result == nil {
		return ErrNilResult
	}

	return nil
}
