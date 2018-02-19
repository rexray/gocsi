#!/bin/sh

HOME=${HOME:-/tmp}
GOPATH=${GOPATH:-$HOME/go}
GOPATH=$(echo "$GOPATH" | awk '{print $1}')

if [ "$1" = "" ]; then
  echo "usage: $0 GO_IMPORT_PATH"
  exit 1
fi

SP_PATH=$1
SP_DIR=$GOPATH/src/$SP_PATH
SP_NAME=$(basename "$SP_PATH")

echo "creating project directories:"
echo "  $SP_DIR"
echo "  $SP_DIR/provider"
echo "  $SP_DIR/service"
mkdir -p "$SP_DIR" "$SP_DIR/provider" "$SP_DIR/service"
cd "$SP_DIR" > /dev/null 2>&1 || exit 1

echo "creating project files:"
echo "  $SP_DIR/main.go"
cat << EOF > "main.go"
package main

import (
	"context"

	"github.com/rexray/gocsi"

	"$SP_PATH/provider"
	"$SP_PATH/service"
)

// main is ignored when this package is built as a go plug-in.
func main() {
	gocsi.Run(
		context.Background(),
		service.Name,
		"A description of the SP",
		"",
		provider.New())
}
EOF

echo "  $SP_DIR/provider/provider.go"
cat << EOF > "provider/provider.go"
package provider

import (
	"context"
	"net"

	log "github.com/sirupsen/logrus"
	"github.com/rexray/gocsi"

	"$SP_PATH/service"
)

// New returns a new CSI Storage Plug-in Provider.
func New() gocsi.StoragePluginProvider {
	svc := service.New()
	return &gocsi.StoragePlugin{
		Controller: svc,
		Identity:   svc,
		Node:       svc,

		// BeforeServe allows the SP to participate in the startup
		// sequence. This function is invoked directly before the
		// gRPC server is created, giving the callback the ability to
		// modify the SP's interceptors, server options, or prevent the
		// server from starting by returning a non-nil error.
		BeforeServe: func(
			ctx context.Context,
			sp *gocsi.StoragePlugin,
			lis net.Listener) error {

			log.WithField("service", service.Name).Debug("BeforeServe")
			return nil
		},

		EnvVars: []string{
			// Enable request validation.
			gocsi.EnvVarSpecReqValidation + "=true",

			// Enable serial volume access.
			gocsi.EnvVarSerialVolAccess + "=true",

			// Provide the list of versions supported by this SP. The
			// specified versions will be:
			//     * Returned by GetSupportedVersions
			//     * Used to validate the Version field of incoming RPCs
			gocsi.EnvVarSupportedVersions + "=" + service.SupportedVersions,
		},
	}
}
EOF

echo "  $SP_DIR/service/service.go"
cat << EOF > "service/service.go"
package service

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
)

const (
	// Name is the name of this CSI SP.
	Name = "$SP_NAME"

	// VendorVersion is the version of this CSP SP.
	VendorVersion = "0.0.0"

	// SupportedVersions is a list of the CSI versions this SP supports.
	SupportedVersions = "0.0.0"
)

// Service is a CSI SP and idempotency.Provider.
type Service interface {
	csi.ControllerServer
	csi.IdentityServer
	csi.NodeServer
}

type service struct{}

// New returns a new Service.
func New() Service {
	return &service{}
}
EOF

echo "  $SP_DIR/service/controller.go"
cat << EOF > "service/controller.go"
package service

import (
	"golang.org/x/net/context"

	"github.com/container-storage-interface/spec/lib/go/csi"
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

echo "  $SP_DIR/service/identity.go"
cat << EOF > "service/identity.go"
package service

import (
	"golang.org/x/net/context"

	"github.com/container-storage-interface/spec/lib/go/csi"
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

echo "  $SP_DIR/service/node.go"
cat << EOF > "service/node.go"
package service

import (
	"golang.org/x/net/context"

	"github.com/container-storage-interface/spec/lib/go/csi"
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

# get dep if necessary and then execute "dep init"
dep_init() {
  DEP=${DEP:-$(which dep 2> /dev/null)}
  DEP_LOG=${DEP_LOG:-.dep.log}
  if [ "$DEP" = "" ]; then
    if [ "$GOHOSTOS" = "" ] || [ "$GOHOSTARCH" = "" ]; then
      GOVERSION=${GO_VERSION:-$(go version | awk '{print $4}')}
      GOHOSTOS=${GOHOSTOS:-$(echo "$GOVERSION" | awk -F/ '{print $1}')}
      GOHOSTARCH=${GOHOSTARCH:-$(echo "$GOVERSION" | awk -F/ '{print $2}')}
    fi
    DEP=./dep
    DEP_VER=${DEP_VER:-0.4.1}
    DEP_BIN=${DEP_BIN:-dep-$GOHOSTOS-$GOHOSTARCH}
    DEP_URL=https://github.com/golang/dep/releases/download/v$DEP_VER/$DEP_BIN
    echo "  downloading golang/dep@v$DEP_VER"
    curl -sSLO "$DEP_URL"
    chmod 0755 "$DEP_BIN"
    mv -f "$DEP_BIN" "$DEP"
  fi
  if [ ! -e Gopkg.toml ]; then
    cat << EOF > Gopkg.toml
# Refer to https://github.com/golang/dep/blob/master/docs/Gopkg.toml.md
# for detailed Gopkg.toml documentation.
#
# Refer to https://github.com/toml-lang/toml for detailed TOML docs.

[[constraint]]
  name = "github.com/rexray/gocsi"
EOF

    if [ "$GOCSI_DEP_BRANCH" != "" ]; then
      echo "  branch = \"$GOCSI_DEP_BRANCH\"" >> Gopkg.toml
    elif [ "$GOCSI_DEP_REVISION" != "" ]; then
      echo "  revision = \"$GOCSI_DEP_REVISION\"" >> Gopkg.toml
    elif [ "$GOCSI_DEP_VERSION" != "" ]; then
      echo "  version = \"$GOCSI_DEP_VERSION\"" >> Gopkg.toml
    else
      echo "  branch = \"master\"" >> Gopkg.toml
    fi

    if [ "$GOCSI_DEP_SOURCE" != "" ]; then
      echo "  source = \"$GOCSI_DEP_SOURCE\"" >> Gopkg.toml
    fi

  fi

  echo "  executing dep ensure"
  if ! "$DEP" ensure > "$DEP_LOG" 2>&1; then cat "$DEP_LOG"; fi
  rm -f "$DEP_LOG"
}

if [ "$USE_DEP" = "true" ]; then
  echo "using golang/dep:"
  dep_init
else
  while true; do
    printf "use golang/dep? Enter yes (default) or no and press [ENTER]: "
    read -r A
    if [ "$A" = "" ] || echo "$A" | grep -iq 'y\(es\)\{0,\}'; then
      dep_init
      break
    fi
    if echo "$A" | grep -iq 'n\(o\)\{0,\}'; then
      break
    fi
  done
fi

echo "building $SP_NAME:"
go build .
BUILD_RESULT=$?

cd - > /dev/null 2>&1 || exit 1

if [ "$BUILD_RESULT" -eq 0 ]; then
  echo "  success!"
  echo '  example: CSI_ENDPOINT=csi.sock \'
  echo '           X_CSI_LOG_LEVEL=info \'
  echo "           $SP_DIR/$SP_NAME"
  echo
  echo "  help available online at"
  echo "  https://github.com/rexray/gocsi#bootstrapping-a-storage-plug-in"
else
  exit 1
fi
