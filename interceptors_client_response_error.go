package gocsi

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/thecodeteam/gocsi/csi"
)

// ClientCheckReponseError is a unary, client validator that checks a
// reply's message to see if it contains an error and transforms it
// into an *Error object, which adheres to Go's Error interface.
func ClientCheckReponseError(
	ctx context.Context,
	method string,
	req, rep interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption) error {

	// Invoke the call and check the reply for an error.
	if err := invoker(ctx, method, req, rep, cc, opts...); err != nil {
		return &Error{
			FullMethod: method,
			InnerError: err,
		}
	}

	switch trep := rep.(type) {

	// Controller
	case *csi.CreateVolumeResponse:
		if err := CheckResponseErrCreateVolume(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.DeleteVolumeResponse:
		if err := CheckResponseErrDeleteVolume(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.ControllerPublishVolumeResponse:
		if err := CheckResponseErrControllerPublishVolume(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.ControllerUnpublishVolumeResponse:
		if err := CheckResponseErrControllerUnpublishVolume(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.ValidateVolumeCapabilitiesResponse:
		if err := CheckResponseErrValidateVolumeCapabilities(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.ListVolumesResponse:
		if err := CheckResponseErrListVolumes(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.GetCapacityResponse:
		if err := CheckResponseErrGetCapacity(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.ControllerGetCapabilitiesResponse:
		if err := CheckResponseErrControllerGetCapabilities(
			ctx, method, trep); err != nil {
			return err
		}

	// Identity
	case *csi.GetSupportedVersionsResponse:
		if err := CheckResponseErrGetSupportedVersions(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.GetPluginInfoResponse:
		if err := CheckResponseErrGetPluginInfo(
			ctx, method, trep); err != nil {
			return err
		}

	// Node
	case *csi.NodePublishVolumeResponse:
		if err := CheckResponseErrNodePublishVolume(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.NodeUnpublishVolumeResponse:
		if err := CheckResponseErrNodeUnpublishVolume(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.GetNodeIDResponse:
		if err := CheckResponseErrGetNodeID(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.NodeProbeResponse:
		if err := CheckResponseErrNodeProbe(
			ctx, method, trep); err != nil {
			return err
		}
	case *csi.NodeGetCapabilitiesResponse:
		if err := CheckResponseErrNodeGetCapabilities(
			ctx, method, trep); err != nil {
			return err
		}
	}

	return nil
}
