package provider

import (
	"github.com/thecodeteam/gocsi/csp"
	"github.com/thecodeteam/gocsi/mock/service"
)

// New returns a new Mock Storage Plug-in Provider.
func New() csp.StoragePluginProvider {
	svc := service.New()
	return &csp.StoragePlugin{
		Controller:          svc,
		Identity:            svc,
		Node:                svc,
		IdempotencyProvider: svc,
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
