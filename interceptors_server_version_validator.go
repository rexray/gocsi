package gocsi

import (
	"fmt"

	"github.com/thecodeteam/gocsi/csi"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type hasGetVersion interface {
	GetVersion() *csi.Version
}

// NewServerRequestVersionValidator initializes a new unary server
// interceptor that validates request versions against the list of
// supported versions.
func NewServerRequestVersionValidator(
	supported []*csi.Version) grpc.UnaryServerInterceptor {

	return (&serverReqVersionValidator{
		supported: supported,
	}).handle
}

type serverReqVersionValidator struct {
	supported []*csi.Version
}

func (v *serverReqVersionValidator) handle(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	// Skip version validation if no supported versions are provided.
	if len(v.supported) == 0 {
		return handler(ctx, req)
	}

	treq, ok := req.(hasGetVersion)
	if !ok {
		return handler(ctx, req)
	}

	rv := treq.GetVersion()

	for _, sv := range v.supported {
		if CompareVersions(rv, sv) == 0 {
			return handler(ctx, req)
		}
	}

	msg := fmt.Sprintf(
		"unsupported request version: %s", SprintfVersion(rv))

	switch req.(type) {
	case *csi.ControllerGetCapabilitiesRequest:
		return ErrControllerGetCapabilities(
			csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION, msg), nil
	case *csi.ControllerPublishVolumeRequest:
		return ErrControllerPublishVolumeGeneral(
			csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION, msg), nil
	case *csi.ControllerUnpublishVolumeRequest:
		return ErrControllerUnpublishVolumeGeneral(
			csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION, msg), nil
	case *csi.CreateVolumeRequest:
		return ErrCreateVolumeGeneral(
			csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION, msg), nil
	case *csi.DeleteVolumeRequest:
		return ErrDeleteVolumeGeneral(
			csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION, msg), nil
	case *csi.GetCapacityRequest:
		return ErrGetCapacity(
			csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION, msg), nil
	case *csi.GetNodeIDRequest:
		return ErrGetNodeIDGeneral(
			csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION, msg), nil
	case *csi.GetPluginInfoRequest:
		return ErrGetPluginInfo(
			csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION, msg), nil
	case *csi.ListVolumesRequest:
		return ErrListVolumes(
			csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION, msg), nil
	case *csi.GetSupportedVersionsRequest:
		panic("Version Check Unsupported for GetSupportedVersions")
	case *csi.NodeGetCapabilitiesRequest:
		return ErrNodeGetCapabilities(
			csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION, msg), nil
	case *csi.NodePublishVolumeRequest:
		return ErrNodePublishVolumeGeneral(
			csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION, msg), nil
	case *csi.NodeUnpublishVolumeRequest:
		return ErrNodeUnpublishVolumeGeneral(
			csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION, msg), nil
	case *csi.NodeProbeRequest:
		return ErrNodeProbeGeneral(
			csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION, msg), nil
	case *csi.ValidateVolumeCapabilitiesRequest:
		return ErrValidateVolumeCapabilitiesGeneral(
			csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION, msg), nil
	}

	panic("Version Check Unsupported")
}
