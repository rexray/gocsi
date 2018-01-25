package provider

import (
	"context"
	"net"

	log "github.com/sirupsen/logrus"

	"github.com/thecodeteam/gocsi"
	"github.com/thecodeteam/gocsi/mock/service"
)

// New returns a new Mock Storage Plug-in Provider.
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
			// Enable serial volume access. Please note that setting
			// X_CSI_SERIAL_VOL_ACCESS=true does not by itself enable the
			// serial volume access middleware. The storage plug-in's
			// GetVolumeID function must be provided as well.
			gocsi.EnvVarSerialVolAccess + "=true",

			// Treat the following fields as required:
			//    * ControllerPublishVolumeRequest.NodeId
			//    * GetNodeIDResponse.NodeId
			gocsi.EnvVarRequireNodeID + "=true",

			// Treat the following fields as required:
			//    * ControllerPublishVolumeResponse.PublishVolumeInfo
			//    * NodePublishVolumeRequest.PublishVolumeInfo
			gocsi.EnvVarRequirePubVolInfo + "=true",

			// Treat CreateVolume responses as successful
			// when they have an associated error code of AlreadyExists.
			gocsi.EnvVarCreateVolAlreadyExistsSuccess + "=true",

			// Treat DeleteVolume responses as successful
			// when they have an associated error code of NotFound.
			gocsi.EnvVarDeleteVolNotFoundSuccess + "=true",

			// Provide the list of versions supported by this SP. The
			// specified versions will be:
			//     * Returned by GetSupportedVersions
			//     * Used to validate the Version field of incoming RPCs
			gocsi.EnvVarSupportedVersions + "=" + service.SupportedVersions,
		},
	}
}
