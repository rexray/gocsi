package gocsi

import (
	"errors"
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/codedellemc/gocsi/csi"
)

const (
	fmIdentPfx = "/" + Namespace + ".Identity/"

	// FMGetSupportedVersions is the eponymous, full method name.
	FMGetSupportedVersions = fmIdentPfx +
		"GetSupportedVersions"
	// FMGetPluginInfo is the eponymous, full method name.
	FMGetPluginInfo = fmIdentPfx +
		"GetPluginInfo"
)

// GetSupportedVersions issues a
// GetSupportedVersions request
// to a CSI controller.
func GetSupportedVersions(
	ctx context.Context,
	c csi.IdentityClient,
	callOpts ...grpc.CallOption) ([]*csi.Version, error) {

	req := &csi.GetSupportedVersionsRequest{}

	res, err := c.GetSupportedVersions(ctx, req, callOpts...)
	if err != nil {
		return nil, err
	}

	// check to see if there is a csi error
	if cerr := res.GetError(); cerr != nil {
		if err := cerr.GetGeneralError(); err != nil {
			return nil, fmt.Errorf(
				"error: GetSupportedVersions failed: %d: %s",
				err.GetErrorCode(),
				err.GetErrorDescription())
		}
		return nil, errors.New(cerr.String())
	}

	result := res.GetResult()
	if result == nil {
		return nil, ErrNilResult
	}

	data := result.GetSupportedVersions()
	if data == nil {
		return nil, ErrNilSupportedVersions
	}

	return data, nil
}

// GetPluginInfo issues a
// GetPluginInfo request
// to a CSI controller.
func GetPluginInfo(
	ctx context.Context,
	c csi.IdentityClient,
	version *csi.Version,
	callOpts ...grpc.CallOption) (*csi.GetPluginInfoResponse_Result, error) {

	if version == nil {
		return nil, ErrVersionRequired
	}

	req := &csi.GetPluginInfoRequest{
		Version: version,
	}

	res, err := c.GetPluginInfo(ctx, req, callOpts...)
	if err != nil {
		return nil, err
	}

	// check to see if there is a csi error
	if cerr := res.GetError(); cerr != nil {
		if err := cerr.GetGeneralError(); err != nil {
			return nil, fmt.Errorf(
				"error: GetPluginInfo failed: %d: %s",
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
