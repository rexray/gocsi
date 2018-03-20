package provider

import (
	"os"

	"github.com/rexray/gocsi"
	"github.com/rexray/gocsi/middleware/logging"
	"github.com/rexray/gocsi/middleware/plugininfo"
	"github.com/rexray/gocsi/middleware/requestid"
	"github.com/rexray/gocsi/middleware/serialvolume"
	"github.com/rexray/gocsi/middleware/specvalidator"
	"github.com/rexray/gocsi/mock/service"
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
		BeforeServe: svc.BeforeServe,

		// Assign the Mock SP's default middleware.
		Middleware: []gocsi.ServerMiddleware{
			// Enable request ID injection.
			&requestid.Middleware{},

			// Enable logging of gRPC request and response data.
			&logging.Middleware{
				RequestWriter:  os.Stdout,
				ResponseWriter: os.Stdout,
			},

			// Enable validation of request or response data using the CSI
			// specification.
			&specvalidator.Middleware{

				// Enable validation of request data.
				RequestValidation: true,

				// Enable validation of response data.
				ResponseValidation: true,

				// Treat the following fields as required:
				//    * ControllerPublishVolumeRequest.NodeId
				//    * NodeGetIdResponse.NodeId
				RequireNodeID: true,

				// Treat the following fields as required:
				//    * ControllerPublishVolumeResponse.PublishInfo
				//    * NodeStageVolumeRequest.PublishInfo
				//    * NodePublishVolumeRequest.PublishInfo
				RequirePublishInfo: true,
			},

			// Enable the runtime assignment of the SP's GetPluginInfo data.
			&plugininfo.Middleware{},

			// Enable serial access to volume resources.
			&serialvolume.Middleware{
				LockProvider: newLockProvider(),
			},
		},
	}
}
