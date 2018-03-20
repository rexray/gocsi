package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/rexray/gocsi/middleware/logging"
	"github.com/rexray/gocsi/middleware/requestid"
	"github.com/rexray/gocsi/middleware/specvalidator"
	"github.com/rexray/gocsi/utils"
)

func getClientMiddleware() grpc.DialOption {
	var interceptors []grpc.UnaryClientInterceptor

	// Configure logging.
	if root.withReqLogging || root.withRepLogging {
		// Automatically enable request ID injection if logging
		// is enabled.
		interceptors = append(
			interceptors, (&requestid.Middleware{}).HandleClient)
		log.Debug("enabled request ID injector")

		mw := logging.Middleware{}
		if root.withReqLogging {
			mw.RequestWriter = os.Stdout
			log.Debug("enabled request logging")
		}
		if root.withRepLogging {
			mw.ResponseWriter = os.Stdout
			log.Debug("enabled response logging")
		}
		interceptors = append(interceptors, mw.HandleClient)
	}

	// Configure the spec validator.
	root.withSpecValidator = root.withSpecValidator ||
		root.withRequiresCreds ||
		root.withRequiresNodeID ||
		root.withRequiresPubVolInfo ||
		root.withRequiresVolumeAttributes
	if root.withSpecValidator {
		mw := specvalidator.Middleware{}
		if root.withRequiresCreds {
			mw.RequireControllerCreateVolumeSecrets = true
			mw.RequireControllerDeleteVolumeSecrets = true
			mw.RequireControllerPublishVolumeSecrets = true
			mw.RequireControllerUnpublishVolumeSecrets = true
			mw.RequireNodeStageVolumeSecrets = true
			mw.RequireNodePublishVolumeSecrets = true
			log.Debug("enabled spec validator opt: requires secrets")
		}
		if root.withRequiresNodeID {
			mw.RequireNodeID = true
			log.Debug("enabled spec validator opt: requires node ID")
		}
		if root.withRequiresPubVolInfo {
			mw.RequirePublishInfo = true
			log.Debug("enabled spec validator opt: requires pub vol info")
		}
		if root.withRequiresVolumeAttributes {
			mw.RequireVolumeAttributes = true
			log.Debug("enabled spec validator opt: requires vol attribs")
		}
		interceptors = append(interceptors, mw.HandleClient)
	}

	if len(interceptors) == 0 {
		return nil
	}

	return grpc.WithUnaryInterceptor(utils.ChainUnaryClient(interceptors...))
}
