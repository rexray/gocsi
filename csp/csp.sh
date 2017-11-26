#!/bin/sh

HOME=${HOME:-/tmp}
GOPATH=${GOPATH:-$HOME/go}
GOPATH=$(echo "$GOPATH" | awk '{print $1}')

if [ "$1" == "" ]; then
  echo "usage: $0 GO_IMPORT_PATH"
  exit 1
fi

SP_PATH=$1
SP_DIR=$GOPATH/src/$SP_PATH
SP_NAME=$(basename "$SP_PATH")

mkdir -p "$SP_DIR" "$SP_DIR/service" "$SP_DIR/provider"

echo "creating $SP_DIR/main.go"
cat << EOF > "$SP_DIR/main.go"
package main

import (
	"context"

	"github.com/thecodeteam/gocsi/csp"

	"$SP_PATH/provider"
	"$SP_PATH/service"
)

// main is ignored when this package is built as a go plug-in.
func main() {
	csp.Run(
		context.Background(),
		service.Name,
		"A description of the SP",
		"",
		provider.New())
}
EOF

echo "creating $SP_DIR/provider/provider.go"
cat << EOF > "$SP_DIR/provider/provider.go"
package provider

import (
	"github.com/thecodeteam/gocsi/csp"

	"$SP_PATH/service"
)

// New returns a new Storage Plug-in Provider.
func New() csp.StoragePluginProvider {
	svc := service.New()
	return &csp.StoragePlugin{
		Controller:          svc,
		Identity:            svc,
		Node:                svc,
		EnvVars: []string{
			// Provide the list of versions supported by this SP. The
			// specified versions will be:
			//     * Returned by GetSupportedVersions
			//     * Used to validate the Version field of incoming RPCs
			csp.EnvVarSupportedVersions + "=" + service.SupportedVersions,
		},
	}
}
EOF

echo "creating $SP_DIR/service/service.go"
cat << EOF > "$SP_DIR/service/service.go"
package service

import (
	"github.com/thecodeteam/gocsi/csi"
)

const (
	// Name is the name of the CSI plug-in.
	Name = "$SP_NAME"

	// VendorVersion is the version returned by GetPluginInfo.
	VendorVersion = "0.0.0"

	// SupportedVersions is a list of supported CSI versions.
	SupportedVersions = "0.0.0"
)

// Service is the CSI service provider.
type Service interface {
	csi.ControllerServer
	csi.IdentityServer
	csi.NodeServer\
}

type service struct {}

// New returns a new Service.
func New() Service {
	return &service{}
}
EOF

echo "creating $SP_DIR/service/controller.go"
cat << EOF > "$SP_DIR/service/controller.go"
package service

import (
	"golang.org/x/net/context"

	"github.com/thecodeteam/gocsi/csi"
)

func (s *service) CreateVolume(
	ctx context.Context,
	req *csi.CreateVolumeRequest) (
	*csi.CreateVolumeResponse, error) {

	return nil, nil
}

func (s *service) DeleteVolume(
	ctx context.Context,
	req *csi.DeleteVolumeRequest) (
	*csi.DeleteVolumeResponse, error) {

	return nil, nil
}

func (s *service) ControllerPublishVolume(
	ctx context.Context,
	req *csi.ControllerPublishVolumeRequest) (
	*csi.ControllerPublishVolumeResponse, error) {

	return nil, nil
}

func (s *service) ControllerUnpublishVolume(
	ctx context.Context,
	req *csi.ControllerUnpublishVolumeRequest) (
	*csi.ControllerUnpublishVolumeResponse, error) {

	return nil, nil
}

func (s *service) ValidateVolumeCapabilities(
	ctx context.Context,
	req *csi.ValidateVolumeCapabilitiesRequest) (
	*csi.ValidateVolumeCapabilitiesResponse, error) {

	return nil, nil
}

func (s *service) ListVolumes(
	ctx context.Context,
	req *csi.ListVolumesRequest) (
	*csi.ListVolumesResponse, error) {

	return nil, nil
}

func (s *service) GetCapacity(
	ctx context.Context,
	req *csi.GetCapacityRequest) (
	*csi.GetCapacityResponse, error) {

	return nil, nil
}

func (s *service) ControllerGetCapabilities(
	ctx context.Context,
	req *csi.ControllerGetCapabilitiesRequest) (
	*csi.ControllerGetCapabilitiesResponse, error) {

	return nil, nil
}

func (s *service) ControllerProbe(
	ctx context.Context,
	req *csi.ControllerProbeRequest) (
	*csi.ControllerProbeResponse, error) {

	return nil, nil
}
EOF

echo "creating $SP_DIR/service/identity.go"
cat << EOF > "$SP_DIR/service/identity.go"
package service

import (
	"golang.org/x/net/context"

	"github.com/thecodeteam/gocsi/csi"
)

func (s *service) GetSupportedVersions(
	ctx context.Context,
	req *csi.GetSupportedVersionsRequest) (
	*csi.GetSupportedVersionsResponse, error) {

	return nil, nil
}

func (s *service) GetPluginInfo(
	ctx context.Context,
	req *csi.GetPluginInfoRequest) (
	*csi.GetPluginInfoResponse, error) {

	return nil, nil
}
EOF

echo "creating $SP_DIR/service/node.go"
cat << EOF > "$SP_DIR/service/node.go"
package service

import (
	"golang.org/x/net/context"

	"github.com/thecodeteam/gocsi/csi"
)

func (s *service) NodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest) (
	*csi.NodePublishVolumeResponse, error) {

	return nil, nil
}

func (s *service) NodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest) (
	*csi.NodeUnpublishVolumeResponse, error) {

	return nil, nil
}

func (s *service) GetNodeID(
	ctx context.Context,
	req *csi.GetNodeIDRequest) (
	*csi.GetNodeIDResponse, error) {

	return nil, nil
}

func (s *service) NodeProbe(
	ctx context.Context,
	req *csi.NodeProbeRequest) (
	*csi.NodeProbeResponse, error) {

	return nil, nil
}

func (s *service) NodeGetCapabilities(
	ctx context.Context,
	req *csi.NodeGetCapabilitiesRequest) (
	*csi.NodeGetCapabilitiesResponse, error) {

	return nil, nil
}
EOF

echo "building $SP_NAME"
go build $SP_PATH/...
