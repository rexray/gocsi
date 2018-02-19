package cmd

import (
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/rexray/gocsi/middleware/logging"
	"github.com/rexray/gocsi/middleware/requestid"
	"github.com/rexray/gocsi/middleware/specvalidator"
	"github.com/rexray/gocsi/utils"
)

func getClientInterceptorsDialOpt() grpc.DialOption {
	var iceptors []grpc.UnaryClientInterceptor

	// Configure logging.
	if root.withReqLogging || root.withRepLogging {

		// Automatically enable request ID injection if logging
		// is enabled.
		iceptors = append(iceptors,
			requestid.NewClientRequestIDInjector())
		log.Debug("enabled request ID injector")

		var (
			loggingOpts []logging.Option
			w           = newLogger(log.Infof)
		)

		if root.withReqLogging {
			loggingOpts = append(loggingOpts, logging.WithRequestLogging(w))
			log.Debug("enabled request logging")
		}
		if root.withRepLogging {
			loggingOpts = append(loggingOpts, logging.WithResponseLogging(w))
			log.Debug("enabled response logging")
		}
		iceptors = append(iceptors,
			logging.NewClientLogger(loggingOpts...))
	}

	// Configure the spec validator.
	root.withSpecValidator = root.withSpecValidator ||
		root.withRequiresCreds ||
		root.withRequiresNodeID ||
		root.withRequiresPubVolInfo ||
		root.withRequiresVolumeAttributes
	if root.withSpecValidator {
		var specOpts []specvalidator.Option
		if root.withRequiresCreds {
			specOpts = append(specOpts,
				specvalidator.WithRequiresCreateVolumeCredentials(),
				specvalidator.WithRequiresDeleteVolumeCredentials(),
				specvalidator.WithRequiresControllerPublishVolumeCredentials(),
				specvalidator.WithRequiresControllerUnpublishVolumeCredentials(),
				specvalidator.WithRequiresNodePublishVolumeCredentials(),
				specvalidator.WithRequiresNodeUnpublishVolumeCredentials())
			log.Debug("enabled spec validator opt: requires creds")
		}
		if root.withRequiresNodeID {
			specOpts = append(specOpts,
				specvalidator.WithRequiresNodeID())
			log.Debug("enabled spec validator opt: requires node ID")
		}
		if root.withRequiresPubVolInfo {
			specOpts = append(specOpts,
				specvalidator.WithRequiresPublishVolumeInfo())
			log.Debug("enabled spec validator opt: requires pub vol info")
		}
		if root.withRequiresVolumeAttributes {
			specOpts = append(specOpts,
				specvalidator.WithRequiresVolumeAttributes())
			log.Debug("enabled spec validator opt: requires vol attribs")
		}
		iceptors = append(iceptors,
			specvalidator.NewClientSpecValidator(specOpts...))
	}

	if len(iceptors) > 0 {
		return grpc.WithUnaryInterceptor(utils.ChainUnaryClient(iceptors...))
	}

	return nil
}
