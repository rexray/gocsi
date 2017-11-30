package provider

import (
	"context"
	"net"

	log "github.com/sirupsen/logrus"

	"github.com/thecodeteam/gocsi/csp"
	"github.com/thecodeteam/gocsi/mock/service"
)

// New returns a new Mock Storage Plug-in Provider.
func New() csp.StoragePluginProvider {
	svc := service.New()
	return &csp.StoragePlugin{
		Controller: svc,
		Identity:   svc,
		Node:       svc,

		// IdempotencyProvider allows an SP to implement idempotency
		// with the most minimal of effort. Please note that providing
		// an IdempotencyProvider does not by itself enable idempotency.
		// The environment variable X_CSI_IDEMP must be set to true as
		// well.
		IdempotencyProvider: svc,

		// BeforeServe allows the SP to participate in the startup
		// sequence. This function is invoked directly before the
		// gRPC server is created, giving the callback the ability to
		// modify the SP's interceptors, server options, or prevent the
		// server from starting by returning a non-nil error.
		BeforeServe: func(
			ctx context.Context,
			sp *csp.StoragePlugin,
			lis net.Listener) error {

			log.WithField("service", service.Name).Debug("BeforeServe")
			return nil
		},

		EnvVars: []string{
			// Enable idempotency. Please note that setting
			// X_CSI_IDEMP=true does not by itself enable the idempotency
			// interceptor. An IdempotencyProvider must be provided as
			// well.
			csp.EnvVarIdemp + "=true",

			// Tell the idempotency interceptor to validate whether or
			// not a volume exists before proceeding with the operation
			csp.EnvVarIdempRequireVolume + "=true",

			// Treat the following fields as required:
			//    * ControllerPublishVolumeRequest.NodeId
			//    * GetNodeIDResponse.NodeId
			csp.EnvVarRequireNodeID + "=true",

			// Treat the following fields as required:
			//    * ControllerPublishVolumeResponse.PublishVolumeInfo
			//    * NodePublishVolumeRequest.PublishVolumeInfo
			csp.EnvVarRequirePubVolInfo + "=true",

			// Treat CreateVolume responses as successful
			// when they have an associated error code of AlreadyExists.
			csp.EnvVarCreateVolAlreadyExistsSuccess + "=true",

			// Treat DeleteVolume responses as successful
			// when they have an associated error code of NotFound.
			csp.EnvVarDeleteVolNotFoundSuccess + "=true",

			// Provide the list of versions supported by this SP. The
			// specified versions will be:
			//     * Returned by GetSupportedVersions
			//     * Used to validate the Version field of incoming RPCs
			csp.EnvVarSupportedVersions + "=" + service.SupportedVersions,
		},
	}
}
