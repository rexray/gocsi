package gocsi

import (
	"fmt"

	"github.com/thecodeteam/gocsi/csi"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// ServerSpecValidator provides a UnaryServerInterceptor that validates
// server request and response data against the CSI specification.
type serverSpecValidator struct {
	opts map[ServerSpecValidatorOption]bool
}

// ServerSpecValidatorOption is an option type used with NewServerSpecValidator.
type ServerSpecValidatorOption uint8

const (
	// NodeIDRequired indicates ControllerPublishVolume requests and
	// GetNodeID responses should contain non-empty node ID data.
	NodeIDRequired ServerSpecValidatorOption = iota

	// PublishVolumeInfoRequired indicates NodePublishVolume requests and
	// ControllerPublishVolume responses should contain non-empty
	// PublishVolumeInfo data.
	PublishVolumeInfoRequired

	// VolumeAttributesRequired indicates ControllerPublishVolume,
	// ValidateVolumeCapabilities, and NodePublishVolume requests
	// should contain non-empty volume attribute data.
	VolumeAttributesRequired

	// CreateVolumeCredentialsRequired indicates CreateVolume requests
	// should contain non-empty credential data.
	CreateVolumeCredentialsRequired

	// DeleteVolumeCredentialsRequired indicates DeleteVolume requests
	// should contain non-empty credential data.
	DeleteVolumeCredentialsRequired

	// ControllerPublishVolumeCredentialsRequired indicates
	// ControllerPublishVolume requests should contain non-empty
	// credential data.
	ControllerPublishVolumeCredentialsRequired

	// ControllerUnpublishVolumeCredentialsRequired indicates
	// ControllerUnpublishVolume requests should contain non-empty
	// credential data.
	ControllerUnpublishVolumeCredentialsRequired

	// NodePublishVolumeCredentialsRequired indicates indicates
	// NodePublishVolume requests should contain non-empty
	// credential data.
	NodePublishVolumeCredentialsRequired

	// NodeUnpublishVolumeCredentialsRequired indicates indicates
	// NodeUnpublishVolume requests should contain non-empty
	// credential data.
	NodeUnpublishVolumeCredentialsRequired
)

// NewServerSpecValidator returns a new UnaryServerInterceptor that validates
// server request and response data against the CSI specification.
func NewServerSpecValidator(
	opts ...ServerSpecValidatorOption) grpc.UnaryServerInterceptor {

	i := &serverSpecValidator{}
	if len(opts) > 0 {
		i.opts = map[ServerSpecValidatorOption]bool{}
		for _, o := range opts {
			i.opts[o] = true
		}
	}
	return i.handle
}

// Handle may be used as a UnaryServerInterceptor to validate incoming
// server-side request data.
func (s *serverSpecValidator) handle(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	switch treq := req.(type) {

	// Controller
	case *csi.CreateVolumeRequest:
		rep, err := s.createVolume(ctx, treq)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			return rep, nil
		}
	case *csi.DeleteVolumeRequest:
		rep, err := s.deleteVolume(ctx, treq)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			return rep, nil
		}
	case *csi.ControllerPublishVolumeRequest:
		rep, err := s.controllerPublishVolume(ctx, treq)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			return rep, nil
		}
	case *csi.ControllerUnpublishVolumeRequest:
		rep, err := s.controllerUnpublishVolume(ctx, treq)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			return rep, nil
		}
	case *csi.ValidateVolumeCapabilitiesRequest:
		rep, err := s.validateVolumeCapabilities(ctx, treq)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			return rep, nil
		}
	case *csi.ListVolumesRequest:
		rep, err := s.listVolumes(ctx, treq)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			return rep, nil
		}
	case *csi.GetCapacityRequest:
		rep, err := s.getCapacity(ctx, treq)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			return rep, nil
		}
	case *csi.ControllerGetCapabilitiesRequest:
		rep, err := s.controllerGetCapabilities(ctx, treq)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			return rep, nil
		}

	// Identity
	case *csi.GetSupportedVersionsRequest:
		rep, err := s.getSupportedVersions(ctx, treq)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			return rep, nil
		}
	case *csi.GetPluginInfoRequest:
		rep, err := s.getPluginInfo(ctx, treq)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			return rep, nil
		}

	// Node
	case *csi.NodePublishVolumeRequest:
		rep, err := s.nodePublishVolume(ctx, treq)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			return rep, nil
		}
	case *csi.NodeUnpublishVolumeRequest:
		rep, err := s.nodeUnpublishVolume(ctx, treq)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			return rep, nil
		}
	case *csi.GetNodeIDRequest:
		rep, err := s.getNodeID(ctx, treq)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			return rep, nil
		}
	case *csi.NodeProbeRequest:
		rep, err := s.nodeProbe(ctx, treq)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			return rep, nil
		}
	case *csi.NodeGetCapabilitiesRequest:
		rep, err := s.nodeGetCapabilities(ctx, treq)
		if err != nil {
			return nil, err
		}
		if rep != nil {
			return rep, nil
		}
	}

	// Call the next chained handler.
	rep, err := handler(ctx, req)
	if err != nil {
		return rep, err
	}

	// If neither of the options are required that involve response
	// validation then jump out early.
	if !(s.opts[NodeIDRequired] || s.opts[PublishVolumeInfoRequired]) {
		return rep, err
	}

	// Do not validate the response if it contains an error.
	if trep, ok := rep.(hasGetError); ok && trep.GetError() != nil {
		return rep, nil
	}

	switch trep := rep.(type) {
	// Controller
	case *csi.ControllerPublishVolumeResponse:
		if !s.opts[PublishVolumeInfoRequired] {
			break
		}
		result := trep.GetResult()
		if result == nil || len(result.PublishVolumeInfo) == 0 {
			return ErrControllerPublishVolumeGeneral(
				csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
				"publish volume info required"), nil
		}

	// Node
	case *csi.GetNodeIDResponse:
		if !s.opts[NodeIDRequired] {
			break
		}
		result := trep.GetResult()
		if result == nil || result.NodeId == "" {
			return ErrGetNodeIDGeneral(
				csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
				"node ID required"), nil
		}
	}

	return rep, err
}

////////////////////////////////////////////////////////////////////////////////
//                      SERVER REQUEST - CONTROLLER                           //
////////////////////////////////////////////////////////////////////////////////

func (s *serverSpecValidator) createVolume(
	ctx context.Context,
	req *csi.CreateVolumeRequest) (
	*csi.CreateVolumeResponse, error) {

	if req.Name == "" {
		return ErrCreateVolume(
			csi.Error_CreateVolumeError_INVALID_VOLUME_NAME,
			"name required"), nil
	}

	if len(req.VolumeCapabilities) == 0 {
		return ErrCreateVolumeGeneral(
			csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
			"volume capabilities required"), nil
	}

	for i, cap := range req.VolumeCapabilities {
		if cap.AccessMode == nil {
			return ErrCreateVolumeGeneral(
				csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
				fmt.Sprintf("access mode required: index %d", i)), nil
		}
		atype := cap.GetAccessType()
		if atype == nil {
			return ErrCreateVolumeGeneral(
				csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
				fmt.Sprintf("access type: index %d required", i)), nil
		}
		switch tatype := atype.(type) {
		case *csi.VolumeCapability_Block:
			if tatype.Block == nil {
				return ErrCreateVolumeGeneral(
					csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
					fmt.Sprintf("block type: index %d required", i)), nil
			}
		case *csi.VolumeCapability_Mount:
			if tatype.Mount == nil {
				return ErrCreateVolumeGeneral(
					csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
					fmt.Sprintf("mount type: index %d required", i)), nil
			}
		default:
			return ErrCreateVolume(
				csi.Error_CreateVolumeError_UNKNOWN,
				fmt.Sprintf(
					"invalid access type: index %d, type=%T",
					i, atype)), nil
		}
	}

	if s.opts[CreateVolumeCredentialsRequired] &&
		len(req.UserCredentials) == 0 {

		return ErrCreateVolumeGeneral(
			csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
			"user credentials required"), nil
	}

	return nil, nil
}

func (s *serverSpecValidator) deleteVolume(
	ctx context.Context,
	req *csi.DeleteVolumeRequest) (
	*csi.DeleteVolumeResponse, error) {

	if req.VolumeId == "" {
		return ErrDeleteVolume(
			csi.Error_DeleteVolumeError_INVALID_VOLUME_ID,
			"volume id required"), nil
	}

	if s.opts[DeleteVolumeCredentialsRequired] &&
		len(req.UserCredentials) == 0 {

		return ErrDeleteVolumeGeneral(
			csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
			"user credentials required"), nil
	}

	return nil, nil
}

func (s *serverSpecValidator) controllerPublishVolume(
	ctx context.Context,
	req *csi.ControllerPublishVolumeRequest) (
	*csi.ControllerPublishVolumeResponse, error) {

	if req.VolumeId == "" {
		return ErrControllerPublishVolume(
			csi.Error_ControllerPublishVolumeError_INVALID_VOLUME_ID,
			"volume id required"), nil
	}

	if s.opts[VolumeAttributesRequired] && len(req.VolumeAttributes) == 0 {
		return ErrControllerPublishVolumeGeneral(
			csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
			"volume attributes required"), nil
	}

	if s.opts[NodeIDRequired] && req.NodeId == "" {
		return ErrControllerPublishVolume(
			csi.Error_ControllerPublishVolumeError_INVALID_NODE_ID,
			"node id required"), nil
	}

	if req.VolumeCapability == nil {
		return ErrControllerPublishVolumeGeneral(
			csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
			"volume capability required"), nil
	}

	if req.VolumeCapability.AccessMode == nil {
		return ErrControllerPublishVolumeGeneral(
			csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
			"access mode required"), nil
	}
	atype := req.VolumeCapability.GetAccessType()
	if atype == nil {
		return ErrControllerPublishVolumeGeneral(
			csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
			"access type required"), nil
	}
	switch tatype := atype.(type) {
	case *csi.VolumeCapability_Block:
		if tatype.Block == nil {
			return ErrControllerPublishVolumeGeneral(
				csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
				"block type required"), nil
		}
	case *csi.VolumeCapability_Mount:
		if tatype.Mount == nil {
			return ErrControllerPublishVolumeGeneral(
				csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
				"mount type required"), nil
		}
	default:
		return ErrControllerPublishVolume(
			csi.Error_ControllerPublishVolumeError_UNKNOWN,
			fmt.Sprintf("invalid access type: %T", atype)), nil
	}

	if s.opts[ControllerPublishVolumeCredentialsRequired] &&
		len(req.UserCredentials) == 0 {

		return ErrControllerPublishVolumeGeneral(
			csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
			"user credentials required"), nil
	}

	return nil, nil
}

func (s *serverSpecValidator) controllerUnpublishVolume(
	ctx context.Context,
	req *csi.ControllerUnpublishVolumeRequest) (
	*csi.ControllerUnpublishVolumeResponse, error) {

	if req.VolumeId == "" {
		return ErrControllerUnpublishVolume(
			csi.Error_ControllerUnpublishVolumeError_INVALID_VOLUME_ID,
			"volume id required"), nil
	}

	if s.opts[ControllerUnpublishVolumeCredentialsRequired] &&
		len(req.UserCredentials) == 0 {

		return ErrControllerUnpublishVolumeGeneral(
			csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
			"user credentials required"), nil
	}

	return nil, nil
}

func (s *serverSpecValidator) validateVolumeCapabilities(
	ctx context.Context,
	req *csi.ValidateVolumeCapabilitiesRequest) (
	*csi.ValidateVolumeCapabilitiesResponse, error) {

	if req.VolumeId == "" {
		return ErrValidateVolumeCapabilities(
			csi.Error_ValidateVolumeCapabilitiesError_INVALID_VOLUME_ID,
			"volume id required"), nil
	}

	if s.opts[VolumeAttributesRequired] && len(req.VolumeAttributes) == 0 {
		return ErrValidateVolumeCapabilitiesGeneral(
			csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
			"volume attributes required"), nil
	}

	if len(req.VolumeCapabilities) == 0 {
		return ErrValidateVolumeCapabilitiesGeneral(
			csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
			"volume capabilities required"), nil
	}

	for i, cap := range req.VolumeCapabilities {
		if cap.AccessMode == nil {
			return ErrValidateVolumeCapabilitiesGeneral(
				csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
				fmt.Sprintf("access mode required: index %d", i)), nil
		}
		atype := cap.GetAccessType()
		if atype == nil {
			return ErrValidateVolumeCapabilitiesGeneral(
				csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
				fmt.Sprintf("access type: index %d required", i)), nil
		}
		switch tatype := atype.(type) {
		case *csi.VolumeCapability_Block:
			if tatype.Block == nil {
				return ErrValidateVolumeCapabilitiesGeneral(
					csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
					fmt.Sprintf("block type: index %d required", i)), nil
			}
		case *csi.VolumeCapability_Mount:
			if tatype.Mount == nil {
				return ErrValidateVolumeCapabilitiesGeneral(
					csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
					fmt.Sprintf("mount type: index %d required", i)), nil
			}
		default:
			return ErrValidateVolumeCapabilities(
				csi.Error_ValidateVolumeCapabilitiesError_UNKNOWN,
				fmt.Sprintf(
					"invalid access type: index %d, type=%T",
					i, atype)), nil
		}
	}

	return nil, nil
}

func (s *serverSpecValidator) listVolumes(
	ctx context.Context,
	req *csi.ListVolumesRequest) (
	*csi.ListVolumesResponse, error) {

	return nil, nil
}

func (s *serverSpecValidator) getCapacity(
	ctx context.Context,
	req *csi.GetCapacityRequest) (
	*csi.GetCapacityResponse, error) {

	if len(req.VolumeCapabilities) == 0 {
		return nil, nil
	}

	for i, cap := range req.VolumeCapabilities {
		if cap.AccessMode == nil {
			return ErrGetCapacity(
				csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
				fmt.Sprintf("access mode required: index %d", i)), nil
		}
		atype := cap.GetAccessType()
		if atype == nil {
			return ErrGetCapacity(
				csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
				fmt.Sprintf("access type: index %d required", i)), nil
		}
		switch tatype := atype.(type) {
		case *csi.VolumeCapability_Block:
			if tatype.Block == nil {
				return ErrGetCapacity(
					csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
					fmt.Sprintf("block type: index %d required", i)), nil
			}
		case *csi.VolumeCapability_Mount:
			if tatype.Mount == nil {
				return ErrGetCapacity(
					csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
					fmt.Sprintf("mount type: index %d required", i)), nil
			}
		default:
			return ErrGetCapacity(
				csi.Error_GeneralError_UNDEFINED,
				fmt.Sprintf(
					"invalid access type: index %d, type=%T",
					i, atype)), nil
		}
	}

	return nil, nil
}

func (s *serverSpecValidator) controllerGetCapabilities(
	ctx context.Context,
	req *csi.ControllerGetCapabilitiesRequest) (
	*csi.ControllerGetCapabilitiesResponse, error) {

	return nil, nil
}

////////////////////////////////////////////////////////////////////////////////
//                        SERVER REQUEST - IDENTITY                           //
////////////////////////////////////////////////////////////////////////////////

func (s *serverSpecValidator) getSupportedVersions(
	ctx context.Context,
	req *csi.GetSupportedVersionsRequest) (
	*csi.GetSupportedVersionsResponse, error) {

	return nil, nil
}

func (s *serverSpecValidator) getPluginInfo(
	ctx context.Context,
	req *csi.GetPluginInfoRequest) (
	*csi.GetPluginInfoResponse, error) {

	return nil, nil
}

////////////////////////////////////////////////////////////////////////////////
//                         SERVER REQUEST - NODE                              //
////////////////////////////////////////////////////////////////////////////////

func (s *serverSpecValidator) nodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest) (
	*csi.NodePublishVolumeResponse, error) {

	if req.VolumeId == "" {
		return ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_INVALID_VOLUME_ID,
			"volume id required"), nil
	}

	if s.opts[VolumeAttributesRequired] && len(req.VolumeAttributes) == 0 {
		return ErrNodePublishVolumeGeneral(
			csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
			"volume attributes required"), nil
	}

	if s.opts[PublishVolumeInfoRequired] && len(req.PublishVolumeInfo) == 0 {
		return ErrNodePublishVolumeGeneral(
			csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
			"publish volume info required"), nil
	}

	if req.VolumeCapability == nil {
		return ErrNodePublishVolumeGeneral(
			csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
			"volume capability required"), nil
	}

	if req.VolumeCapability.AccessMode == nil {
		return ErrNodePublishVolumeGeneral(
			csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
			"access mode required"), nil
	}
	atype := req.VolumeCapability.GetAccessType()
	if atype == nil {
		return ErrNodePublishVolumeGeneral(
			csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
			"access type required"), nil
	}
	switch tatype := atype.(type) {
	case *csi.VolumeCapability_Block:
		if tatype.Block == nil {
			return ErrNodePublishVolumeGeneral(
				csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
				"block type required"), nil
		}
	case *csi.VolumeCapability_Mount:
		if tatype.Mount == nil {
			return ErrNodePublishVolumeGeneral(
				csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
				"mount type required"), nil
		}
	default:
		return ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_UNKNOWN,
			fmt.Sprintf("invalid access type: %T", atype)), nil
	}

	if req.TargetPath == "" {
		return ErrNodePublishVolumeGeneral(
			csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
			"target path required"), nil
	}

	if s.opts[NodePublishVolumeCredentialsRequired] &&
		len(req.UserCredentials) == 0 {

		return ErrNodePublishVolumeGeneral(
			csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
			"user credentials required"), nil
	}

	return nil, nil
}

func (s *serverSpecValidator) nodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest) (
	*csi.NodeUnpublishVolumeResponse, error) {

	if req.VolumeId == "" {
		return ErrNodeUnpublishVolume(
			csi.Error_NodeUnpublishVolumeError_INVALID_VOLUME_ID,
			"volume id required"), nil
	}

	if req.TargetPath == "" {
		return ErrNodeUnpublishVolumeGeneral(
			csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
			"target path required"), nil
	}

	if s.opts[NodeUnpublishVolumeCredentialsRequired] &&
		len(req.UserCredentials) == 0 {

		return ErrNodeUnpublishVolumeGeneral(
			csi.Error_GeneralError_MISSING_REQUIRED_FIELD,
			"user credentials required"), nil
	}

	return nil, nil
}

func (s *serverSpecValidator) getNodeID(
	ctx context.Context,
	req *csi.GetNodeIDRequest) (
	*csi.GetNodeIDResponse, error) {

	return nil, nil
}

func (s *serverSpecValidator) nodeProbe(
	ctx context.Context,
	req *csi.NodeProbeRequest) (
	*csi.NodeProbeResponse, error) {

	return nil, nil
}

func (s *serverSpecValidator) nodeGetCapabilities(
	ctx context.Context,
	req *csi.NodeGetCapabilitiesRequest) (
	*csi.NodeGetCapabilitiesResponse, error) {

	return nil, nil
}
