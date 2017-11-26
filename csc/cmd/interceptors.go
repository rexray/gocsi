package cmd

import (
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/thecodeteam/gocsi"
)

func getClientInterceptorsDialOpt() grpc.DialOption {
	var iceptors []grpc.UnaryClientInterceptor

	// Configure logging.
	if root.withReqLogging || root.withRepLogging {

		// Automatically enable request ID injection if logging
		// is enabled.
		iceptors = append(iceptors,
			gocsi.NewClientRequestIDInjector())
		log.Debug("enabled request ID injector")

		var (
			loggingOpts []gocsi.LoggingOption
			w           = newLogger(log.Infof)
		)

		if root.withReqLogging {
			loggingOpts = append(loggingOpts, gocsi.WithRequestLogging(w))
			log.Debug("enabled request logging")
		}
		if root.withRepLogging {
			loggingOpts = append(loggingOpts, gocsi.WithResponseLogging(w))
			log.Debug("enabled response logging")
		}
		iceptors = append(iceptors,
			gocsi.NewClientLogger(loggingOpts...))
	}

	// Configure the spec validator.
	root.withSpecValidator = root.withSpecValidator ||
		root.withRequiresCreds ||
		root.withSuccessCreateVolumeAlreadyExists ||
		root.withSuccessDeleteVolumeNotFound ||
		root.withRequiresNodeID ||
		root.withRequiresPubVolInfo ||
		root.withRequiresVolumeAttributes
	if root.withSpecValidator {
		var specOpts []gocsi.SpecValidatorOption
		if root.withRequiresCreds {
			specOpts = append(specOpts,
				gocsi.WithRequiresCreateVolumeCredentials(),
				gocsi.WithRequiresDeleteVolumeCredentials(),
				gocsi.WithRequiresControllerPublishVolumeCredentials(),
				gocsi.WithRequiresControllerUnpublishVolumeCredentials(),
				gocsi.WithRequiresNodePublishVolumeCredentials(),
				gocsi.WithRequiresNodeUnpublishVolumeCredentials())
			log.Debug("enabled spec validator opt: requires creds")
		}
		if root.withRequiresNodeID {
			specOpts = append(specOpts,
				gocsi.WithRequiresNodeID())
			log.Debug("enabled spec validator opt: requires node ID")
		}
		if root.withRequiresPubVolInfo {
			specOpts = append(specOpts,
				gocsi.WithRequiresPublishVolumeInfo())
			log.Debug("enabled spec validator opt: requires pub vol info")
		}
		if root.withRequiresVolumeAttributes {
			specOpts = append(specOpts,
				gocsi.WithRequiresVolumeAttributes())
			log.Debug("enabled spec validator opt: requires vol attribs")
		}
		if root.withSuccessCreateVolumeAlreadyExists {
			specOpts = append(specOpts,
				gocsi.WithSuccessCreateVolumeAlreadyExists())
			log.Debug("enabled spec validator opt: create exists success")
		}
		if root.withSuccessDeleteVolumeNotFound {
			specOpts = append(specOpts,
				gocsi.WithSuccessDeleteVolumeNotFound())
			log.Debug("enabled spec validator opt: delete !exists success")
		}
		iceptors = append(iceptors,
			gocsi.NewClientSpecValidator(specOpts...))
	}

	if len(iceptors) > 0 {
		return grpc.WithUnaryInterceptor(gocsi.ChainUnaryClient(iceptors...))
	}

	return nil
}
