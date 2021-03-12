package cmd

import (
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/dell/gocsi/middleware/logging"
	"github.com/dell/gocsi/middleware/requestid"
	"github.com/dell/gocsi/middleware/specvalidator"
	"github.com/dell/gocsi/utils"
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
		root.withRequiresVolContext ||
		root.withRequiresPubContext
	if root.withSpecValidator {
		var specOpts []specvalidator.Option
		if root.withRequiresCreds {
			specOpts = append(specOpts,
				specvalidator.WithRequiresControllerCreateVolumeSecrets(),
				specvalidator.WithRequiresControllerDeleteVolumeSecrets(),
				specvalidator.WithRequiresControllerPublishVolumeSecrets(),
				specvalidator.WithRequiresControllerUnpublishVolumeSecrets(),
				specvalidator.WithRequiresNodeStageVolumeSecrets(),
				specvalidator.WithRequiresNodePublishVolumeSecrets())
			log.Debug("enabled spec validator opt: requires creds")
		}
		if root.withRequiresVolContext {
			specOpts = append(specOpts,
				specvalidator.WithRequiresVolumeContext())
			log.Debug("enabled spec validator opt: requires vol context")
		}
		if root.withRequiresPubContext {
			specOpts = append(specOpts,
				specvalidator.WithRequiresPublishContext())
			log.Debug("enabled spec validator opt: requires pub context")
		}
		iceptors = append(iceptors,
			specvalidator.NewClientSpecValidator(specOpts...))
	}

	if len(iceptors) > 0 {
		return grpc.WithUnaryInterceptor(utils.ChainUnaryClient(iceptors...))
	}

	return nil
}
