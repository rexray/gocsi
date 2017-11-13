package gocsi

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/thecodeteam/gocsi/csi"
)

const (
	// FMGetSupportedVersions is the full method name for the
	// eponymous RPC message.
	FMGetSupportedVersions = "/" + Namespace +
		".Identity/" +
		"GetSupportedVersions"

	// FMGetPluginInfo is the full method name for the
	// eponymous RPC message.
	FMGetPluginInfo = "/" + Namespace +
		".Identity/" +
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
	if res != nil {
		return res.SupportedVersions, err
	}
	return nil, err
}

// GetPluginInfo issues a
// GetPluginInfo request
// to a CSI controller.
func GetPluginInfo(
	ctx context.Context,
	c csi.IdentityClient,
	version *csi.Version,
	callOpts ...grpc.CallOption) (string, string, map[string]string, error) {

	req := &csi.GetPluginInfoRequest{
		Version: version,
	}

	res, err := c.GetPluginInfo(ctx, req, callOpts...)
	if res != nil {
		return res.Name, res.VendorVersion, res.Manifest, err
	}
	return "", "", nil, err
}
