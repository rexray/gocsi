package gocsi

import (
	"bytes"
	"fmt"
	"io"

	"github.com/thecodeteam/gocsi/csi"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// NewServerResponseLogger initializes a new unary, server interceptor
// that logs reply details.
func NewServerResponseLogger(
	stdout, stderr io.Writer) grpc.UnaryServerInterceptor {

	return (&serverRepLogger{stdout: stdout, stderr: stderr}).handle
}

type serverRepLogger struct {
	stdout io.Writer
	stderr io.Writer
}

func (v *serverRepLogger) handle(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	if req == nil {
		return handler(ctx, req)
	}

	w := v.stdout
	b := &bytes.Buffer{}

	fmt.Fprintf(b, "%s: ", info.FullMethod)
	if rid, ok := GetRequestID(ctx); ok {
		fmt.Fprintf(b, "REP %04d", rid)
	}

	rep, err := handler(ctx, req)
	if err != nil {
		fmt.Fprintf(b, ": %v", &Error{
			FullMethod: info.FullMethod,
			Code:       ErrorNoCode,
			InnerError: err,
		})
		fmt.Fprintln(w, b.String())
		return rep, err
	}

	if rep == nil {
		return nil, nil
	}

	var gocsiErr error

	switch trep := rep.(type) {

	// Controller
	case *csi.CreateVolumeResponse:
		gocsiErr = CheckResponseErrCreateVolume(
			ctx, info.FullMethod, trep)
	case *csi.DeleteVolumeResponse:
		gocsiErr = CheckResponseErrDeleteVolume(
			ctx, info.FullMethod, trep)
	case *csi.ControllerPublishVolumeResponse:
		gocsiErr = CheckResponseErrControllerPublishVolume(
			ctx, info.FullMethod, trep)
	case *csi.ControllerUnpublishVolumeResponse:
		gocsiErr = CheckResponseErrControllerUnpublishVolume(
			ctx, info.FullMethod, trep)
	case *csi.ValidateVolumeCapabilitiesResponse:
		gocsiErr = CheckResponseErrValidateVolumeCapabilities(
			ctx, info.FullMethod, trep)
	case *csi.ListVolumesResponse:
		gocsiErr = CheckResponseErrListVolumes(
			ctx, info.FullMethod, trep)
	case *csi.GetCapacityResponse:
		gocsiErr = CheckResponseErrGetCapacity(
			ctx, info.FullMethod, trep)
	case *csi.ControllerGetCapabilitiesResponse:
		gocsiErr = CheckResponseErrControllerGetCapabilities(
			ctx, info.FullMethod, trep)

	// Identity
	case *csi.GetSupportedVersionsResponse:
		gocsiErr = CheckResponseErrGetSupportedVersions(
			ctx, info.FullMethod, trep)
	case *csi.GetPluginInfoResponse:
		gocsiErr = CheckResponseErrGetPluginInfo(
			ctx, info.FullMethod, trep)

	// Node
	case *csi.NodePublishVolumeResponse:
		gocsiErr = CheckResponseErrNodePublishVolume(
			ctx, info.FullMethod, trep)
	case *csi.NodeUnpublishVolumeResponse:
		gocsiErr = CheckResponseErrNodeUnpublishVolume(
			ctx, info.FullMethod, trep)
	case *csi.GetNodeIDResponse:
		gocsiErr = CheckResponseErrGetNodeID(
			ctx, info.FullMethod, trep)
	case *csi.NodeProbeResponse:
		gocsiErr = CheckResponseErrNodeProbe(
			ctx, info.FullMethod, trep)
	case *csi.NodeGetCapabilitiesResponse:
		gocsiErr = CheckResponseErrNodeGetCapabilities(
			ctx, info.FullMethod, trep)
	}

	// Check to see if the reply has an error or is an error itself.
	if gocsiErr != nil {
		fmt.Fprintf(b, ": %v", gocsiErr)
		fmt.Fprintln(w, b.String())
		return rep, err
	}

	// At this point the reply must be valid. Format and print it.
	rprintReqOrRep(b, rep)
	fmt.Fprintln(w, b.String())
	return rep, err
}
