package csp

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/thecodeteam/gocsi"
	"github.com/thecodeteam/gocsi/csi"
)

func (sp *StoragePlugin) injectContext(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	return handler(gocsi.WithLookupEnv(ctx, sp.lookupEnv), req)
}

func (sp *StoragePlugin) getSupportedVersions(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	if info.FullMethod != gocsi.GetSupportedVersions ||
		len(sp.supportedVersions) == 0 {

		return handler(ctx, req)
	}

	rep := &csi.GetSupportedVersionsResponse{
		SupportedVersions: make([]*csi.Version, len(sp.supportedVersions)),
	}
	for i := range sp.supportedVersions {
		rep.SupportedVersions[i] = &sp.supportedVersions[i]
	}

	return rep, nil
}
