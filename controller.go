package gocsi

import (
	"errors"
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/codedellemc/gocsi/csi"
)

// CreateVolume issues a CreateVolume request to a CSI controller.
func CreateVolume(
	ctx context.Context,
	c csi.ControllerClient,
	version *csi.Version,
	name string,
	requiredBytes, limitBytes uint64,
	fsType string, mountFlags []string,
	params map[string]string,
	callOpts ...grpc.CallOption) (volume *csi.VolumeInfo, err error) {

	if version == nil {
		return nil, ErrVersionRequired
	}

	req := &csi.CreateVolumeRequest{
		Name:       name,
		Version:    version,
		Parameters: params,
	}

	if requiredBytes > 0 || limitBytes > 0 {
		req.CapacityRange = &csi.CapacityRange{
			LimitBytes:    limitBytes,
			RequiredBytes: requiredBytes,
		}
	}

	if fsType != "" || len(mountFlags) > 0 {
		cap := &csi.VolumeCapability_MountVolume{}
		cap.FsType = fsType
		if len(mountFlags) > 0 {
			cap.MountFlags = mountFlags
		}
		req.VolumeCapabilities = []*csi.VolumeCapability{
			&csi.VolumeCapability{
				Value: &csi.VolumeCapability_Mount{Mount: cap},
			},
		}
	}

	res, err := c.CreateVolume(ctx, req, callOpts...)
	if err != nil {
		return nil, err
	}

	// check to see if there is a csi error
	if cerr := res.GetError(); cerr != nil {
		if err := cerr.GetCreateVolumeError(); err != nil {
			return nil, fmt.Errorf(
				"error: CreateVolume failed: %d: %s",
				err.GetErrorCode(),
				err.GetErrorDescription())
		}
		if err := cerr.GetGeneralError(); err != nil {
			return nil, fmt.Errorf(
				"error: CreateVolume failed: %d: %s",
				err.GetErrorCode(),
				err.GetErrorDescription())
		}
		return nil, errors.New(cerr.String())
	}

	result := res.GetResult()
	if result == nil {
		return nil, ErrNilResult
	}

	data := result.GetVolumeInfo()
	if data == nil {
		return nil, ErrNilVolumeInfo
	}

	return data, nil
}

// DeleteVolume issues a DeleteVolume request to a CSI controller.
func DeleteVolume(
	ctx context.Context,
	c csi.ControllerClient,
	version *csi.Version,
	volumeID *csi.VolumeID,
	volumeMetadata *csi.VolumeMetadata,
	callOpts ...grpc.CallOption) error {

	if version == nil {
		return ErrVersionRequired
	}

	if volumeID == nil {
		return ErrVolumeIDRequired
	}

	req := &csi.DeleteVolumeRequest{
		Version:        version,
		VolumeId:       volumeID,
		VolumeMetadata: volumeMetadata,
	}

	res, err := c.DeleteVolume(ctx, req, callOpts...)
	if err != nil {
		return err
	}

	// check to see if there is a csi error
	if cerr := res.GetError(); cerr != nil {
		if err := cerr.GetDeleteVolumeError(); err != nil {
			return fmt.Errorf(
				"error: DeleteVolume failed: %d: %s",
				err.GetErrorCode(),
				err.GetErrorDescription())
		}
		if err := cerr.GetGeneralError(); err != nil {
			return fmt.Errorf(
				"error: DeleteVolume failed: %d: %s",
				err.GetErrorCode(),
				err.GetErrorDescription())
		}
		return errors.New(cerr.String())
	}

	return nil
}

// ControllerPublishVolume issues a
// ControllerPublishVolume request
// to a CSI controller.
func ControllerPublishVolume(
	ctx context.Context,
	c csi.ControllerClient,
	version *csi.Version,
	volumeID *csi.VolumeID,
	volumeMetadata *csi.VolumeMetadata,
	nodeID *csi.NodeID,
	readonly bool,
	callOpts ...grpc.CallOption) (
	*csi.PublishVolumeInfo, error) {

	if version == nil {
		return nil, ErrVersionRequired
	}

	if volumeID == nil {
		return nil, ErrVolumeIDRequired
	}

	req := &csi.ControllerPublishVolumeRequest{
		Version:        version,
		VolumeId:       volumeID,
		VolumeMetadata: volumeMetadata,
		NodeId:         nodeID,
		Readonly:       readonly,
	}

	res, err := c.ControllerPublishVolume(ctx, req, callOpts...)
	if err != nil {
		return nil, err
	}

	// check to see if there is a csi error
	if cerr := res.GetError(); cerr != nil {
		if err := cerr.GetControllerPublishVolumeError(); err != nil {
			return nil, fmt.Errorf(
				"error: ControllerPublishVolume failed: %d: %s",
				err.GetErrorCode(),
				err.GetErrorDescription())
		}
		if err := cerr.GetGeneralError(); err != nil {
			return nil, fmt.Errorf(
				"error: ControllerPublishVolume failed: %d: %s",
				err.GetErrorCode(),
				err.GetErrorDescription())
		}
		return nil, errors.New(cerr.String())
	}

	result := res.GetResult()
	if result == nil {
		return nil, ErrNilResult
	}

	data := result.GetPublishVolumeInfo()
	if data == nil {
		return nil, ErrNilPublishVolumeInfo
	}

	return data, nil
}

// ControllerUnpublishVolume issues a
// ControllerUnpublishVolume request
// to a CSI controller.
func ControllerUnpublishVolume(
	ctx context.Context,
	c csi.ControllerClient,
	version *csi.Version,
	volumeID *csi.VolumeID,
	volumeMetadata *csi.VolumeMetadata,
	nodeID *csi.NodeID,
	callOpts ...grpc.CallOption) error {

	if version == nil {
		return ErrVersionRequired
	}

	if volumeID == nil {
		return ErrVolumeIDRequired
	}

	req := &csi.ControllerUnpublishVolumeRequest{
		Version:        version,
		VolumeId:       volumeID,
		VolumeMetadata: volumeMetadata,
		NodeId:         nodeID,
	}

	res, err := c.ControllerUnpublishVolume(ctx, req, callOpts...)
	if err != nil {
		return err
	}

	// check to see if there is a csi error
	if cerr := res.GetError(); cerr != nil {
		if err := cerr.GetControllerUnpublishVolumeError(); err != nil {
			return fmt.Errorf(
				"error: ControllerUnpublishVolume failed: %d: %s",
				err.GetErrorCode(),
				err.GetErrorDescription())
		}
		if err := cerr.GetGeneralError(); err != nil {
			return fmt.Errorf(
				"error: ControllerUnpublishVolume failed: %d: %s",
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

// ValidateVolumeCapabilities issues a ValidateVolumeCapabilities
// request to a CSI controller
func ValidateVolumeCapabilities(
	ctx context.Context,
	c csi.ControllerClient,
	version *csi.Version,
	volumeInfo *csi.VolumeInfo,
	volumeCapabilities []*csi.VolumeCapability,
	callOpts ...grpc.CallOption) (*csi.ValidateVolumeCapabilitiesResponse_Result, error) {

	if version == nil {
		return nil, ErrVersionRequired
	}
	if volumeInfo == nil {
		return nil, ErrVolumeInfoRequired
	}
	if volumeCapabilities == nil {
		return nil, ErrVolumeCapabilityRequired
	}

	req := &csi.ValidateVolumeCapabilitiesRequest{
		Version:            version,
		VolumeInfo:         volumeInfo,
		VolumeCapabilities: volumeCapabilities,
	}

	res, err := c.ValidateVolumeCapabilities(ctx, req, callOpts...)
	if err != nil {
		return nil, err
	}

	// check to see if there is a csi error
	if cerr := res.GetError(); cerr != nil {
		if err := cerr.GetValidateVolumeCapabilitiesError(); err != nil {
			return nil, fmt.Errorf(
				"error: ValidateVolumeCapabilities failed: %d: %s",
				err.GetErrorCode(),
				err.GetErrorDescription())
		}
		if err := cerr.GetGeneralError(); err != nil {
			return nil, fmt.Errorf(
				"error: ValidateVolumeCapabilities failed: %d: %s",
				err.GetErrorCode(),
				err.GetErrorDescription())
		}
		return nil, errors.New(cerr.String())
	}

	result := res.GetResult()
	if result == nil {
		return nil, ErrNilResult
	}

	return result, nil
}

// ListVolumes issues a ListVolumes request to a CSI controller.
func ListVolumes(
	ctx context.Context,
	c csi.ControllerClient,
	version *csi.Version,
	maxEntries uint32,
	startingToken string,
	callOpts ...grpc.CallOption) (
	volumes []*csi.VolumeInfo, nextToken string, err error) {

	if version == nil {
		return nil, "", ErrVersionRequired
	}

	req := &csi.ListVolumesRequest{
		MaxEntries:    maxEntries,
		StartingToken: startingToken,
		Version:       version,
	}

	res, err := c.ListVolumes(ctx, req, callOpts...)
	if err != nil {
		return nil, "", err
	}

	// check to see if there is a csi error
	if cerr := res.GetError(); cerr != nil {
		if err := cerr.GetGeneralError(); err != nil {
			return nil, "", fmt.Errorf(
				"error: ListVolumes failed: %d: %s",
				err.GetErrorCode(),
				err.GetErrorDescription())
		}
		return nil, "", errors.New(cerr.String())
	}

	result := res.GetResult()
	if result == nil {
		return nil, "", ErrNilResult
	}

	nextToken = result.GetNextToken()
	entries := result.GetEntries()

	// check to see if there are zero entries
	if len(entries) == 0 {
		return nil, nextToken, nil
	}

	volumes = make([]*csi.VolumeInfo, len(entries))

	for x, e := range entries {
		if volumes[x] = e.GetVolumeInfo(); volumes[x] == nil {
			return nil, "", ErrNilVolumeInfo
		}
	}

	return volumes, nextToken, nil
}

// GetCapacity issues a GetCapacity
// request to a CSI controller
func GetCapacity(
	ctx context.Context,
	c csi.ControllerClient,
	version *csi.Version,
	callOpts ...grpc.CallOption) (uint64, error) {

	if version == nil {
		return 0, ErrVersionRequired
	}

	req := &csi.GetCapacityRequest{
		Version: version,
	}

	res, err := c.GetCapacity(ctx, req, callOpts...)
	if err != nil {
		return 0, err
	}

	// check to see if there is a csi error
	if cerr := res.GetError(); cerr != nil {
		if err := cerr.GetGeneralError(); err != nil {
			return 0, fmt.Errorf(
				"error: GetCapacity failed: %d: %s",
				err.GetErrorCode(),
				err.GetErrorDescription())
		}
		return 0, errors.New(cerr.String())
	}

	result := res.GetResult()
	if result == nil {
		return 0, ErrNilResult
	}

	cap := result.GetTotalCapacity()

	return cap, nil
}

// ControllerGetCapabilities issues a ControllerGetCapabilities request to a
// CSI controller.
func ControllerGetCapabilities(
	ctx context.Context,
	c csi.ControllerClient,
	version *csi.Version,
	callOpts ...grpc.CallOption) (
	capabilties []*csi.ControllerServiceCapability, err error) {

	if version == nil {
		return nil, ErrVersionRequired
	}

	req := &csi.ControllerGetCapabilitiesRequest{
		Version: version,
	}

	res, err := c.ControllerGetCapabilities(ctx, req, callOpts...)
	if err != nil {
		return nil, err
	}

	// check to see if there is a csi error
	if cerr := res.GetError(); cerr != nil {
		if err := cerr.GetGeneralError(); err != nil {
			return nil, fmt.Errorf(
				"error: ControllerGetCapabilities failed: %d: %s",
				err.GetErrorCode(),
				err.GetErrorDescription())
		}
		return nil, errors.New(cerr.String())
	}

	result := res.GetResult()
	if result == nil {
		return nil, ErrNilResult
	}

	return result.GetCapabilities(), nil
}
