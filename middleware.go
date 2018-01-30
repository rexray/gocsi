package gocsi

import (
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/container-storage-interface/spec/lib/go/csi"

	csictx "github.com/thecodeteam/gocsi/context"
	"github.com/thecodeteam/gocsi/middleware/logging"
	"github.com/thecodeteam/gocsi/middleware/requestid"
	"github.com/thecodeteam/gocsi/middleware/serialvolume"
	"github.com/thecodeteam/gocsi/middleware/serialvolume/etcd"
	"github.com/thecodeteam/gocsi/middleware/specvalidator"
	"github.com/thecodeteam/gocsi/utils"
)

func (sp *StoragePlugin) initInterceptors(ctx context.Context) {

	sp.Interceptors = append(sp.Interceptors, sp.injectContext)
	log.Debug("enabled context injector")

	const (
		envVarNewVolExists   = EnvVarCreateVolAlreadyExistsSuccess
		envVarDelVolNotFound = EnvVarDeleteVolNotFoundSuccess
	)

	var (
		withReqLogging         = sp.getEnvBool(ctx, EnvVarReqLogging)
		withRepLogging         = sp.getEnvBool(ctx, EnvVarRepLogging)
		withSerialVol          = sp.getEnvBool(ctx, EnvVarSerialVolAccess)
		withSpec               = sp.getEnvBool(ctx, EnvVarSpecValidation)
		withNodeID             = sp.getEnvBool(ctx, EnvVarRequireNodeID)
		withPubVolInfo         = sp.getEnvBool(ctx, EnvVarRequirePubVolInfo)
		withVolAttribs         = sp.getEnvBool(ctx, EnvVarRequireVolAttribs)
		withCreds              = sp.getEnvBool(ctx, EnvVarCreds)
		withCredsNewVol        = sp.getEnvBool(ctx, EnvVarCredsCreateVol)
		withCredsDelVol        = sp.getEnvBool(ctx, EnvVarCredsDeleteVol)
		withCredsCtrlrPubVol   = sp.getEnvBool(ctx, EnvVarCredsCtrlrPubVol)
		withCredsCtrlrUnpubVol = sp.getEnvBool(ctx, EnvVarCredsCtrlrUnpubVol)
		withCredsNodePubVol    = sp.getEnvBool(ctx, EnvVarCredsNodePubVol)
		withCredsNodeUnpubVol  = sp.getEnvBool(ctx, EnvVarCredsNodeUnpubVol)
	)

	// Enable all cred requirements if the general option is enabled.
	if withCreds {
		withCredsNewVol = true
		withCredsDelVol = true
		withCredsCtrlrPubVol = true
		withCredsCtrlrUnpubVol = true
		withCredsNodePubVol = true
		withCredsNodeUnpubVol = true
	}

	// Enable spec validation if any of the spec-related options are enabled.
	withSpec = withSpec ||
		withCreds ||
		withNodeID ||
		withPubVolInfo ||
		withVolAttribs

	// Configure logging.
	if withReqLogging || withRepLogging {
		// Automatically enable request ID injection if logging
		// is enabled.
		sp.Interceptors = append(sp.Interceptors,
			requestid.NewServerRequestIDInjector())
		log.Debug("enabled request ID injector")

		var (
			loggingOpts []logging.Option
			w           = newLogger(log.Debugf)
		)

		if withReqLogging {
			loggingOpts = append(loggingOpts, logging.WithRequestLogging(w))
			log.Debug("enabled request logging")
		}
		if withRepLogging {
			loggingOpts = append(loggingOpts, logging.WithResponseLogging(w))
			log.Debug("enabled response logging")
		}
		sp.Interceptors = append(sp.Interceptors,
			logging.NewServerLogger(loggingOpts...))
	}

	if withSpec {
		var specOpts []specvalidator.Option

		if len(sp.supportedVersions) > 0 {
			specOpts = append(
				specOpts,
				specvalidator.WithSupportedVersions(sp.supportedVersions...))
		}
		if withCredsNewVol {
			specOpts = append(specOpts,
				specvalidator.WithRequiresCreateVolumeCredentials())
			log.Debug("enabled spec validator opt: requires creds: " +
				"CreateVolume")
		}
		if withCredsDelVol {
			specOpts = append(specOpts,
				specvalidator.WithRequiresDeleteVolumeCredentials())
			log.Debug("enabled spec validator opt: requires creds: " +
				"DeleteVolume")
		}
		if withCredsCtrlrPubVol {
			specOpts = append(specOpts,
				specvalidator.WithRequiresControllerPublishVolumeCredentials())
			log.Debug("enabled spec validator opt: requires creds: " +
				"ControllerPublishVolume")
		}
		if withCredsCtrlrUnpubVol {
			specOpts = append(specOpts,
				specvalidator.WithRequiresControllerUnpublishVolumeCredentials())
			log.Debug("enabled spec validator opt: requires creds: " +
				"ControllerUnpublishVolume")
		}
		if withCredsNodePubVol {
			specOpts = append(specOpts,
				specvalidator.WithRequiresNodePublishVolumeCredentials())
			log.Debug("enabled spec validator opt: requires creds: " +
				"NodePublishVolume")
		}
		if withCredsNodeUnpubVol {
			specOpts = append(specOpts,
				specvalidator.WithRequiresNodeUnpublishVolumeCredentials())
			log.Debug("enabled spec validator opt: requires creds: " +
				"NodeUnpublishVolume")
		}

		if withNodeID {
			specOpts = append(specOpts,
				specvalidator.WithRequiresNodeID())
			log.Debug("enabled spec validator opt: requires node ID")
		}
		if withPubVolInfo {
			specOpts = append(specOpts,
				specvalidator.WithRequiresPublishVolumeInfo())
			log.Debug("enabled spec validator opt: requires pub vol info")
		}
		if withVolAttribs {
			specOpts = append(specOpts,
				specvalidator.WithRequiresVolumeAttributes())
			log.Debug("enabled spec validator opt: requires vol attribs")
		}
		sp.Interceptors = append(sp.Interceptors,
			specvalidator.NewServerSpecValidator(specOpts...))
	}

	if _, ok := csictx.LookupEnv(ctx, EnvVarPluginInfo); ok {
		log.Debug("enabled GetPluginInfo interceptor")
		sp.Interceptors = append(sp.Interceptors, sp.getPluginInfo)
	}

	if len(sp.supportedVersions) > 0 {
		log.Debug("enabled GetSupportedVersions interceptor")
		sp.Interceptors = append(sp.Interceptors, sp.getSupportedVersions)
	}

	if withSerialVol {
		var (
			opts   []serialvolume.Option
			fields = map[string]interface{}{}
		)

		// Get serial provider's timeout.
		if v, _ := csictx.LookupEnv(
			ctx, EnvVarSerialVolAccessTimeout); v != "" {
			if t, err := time.ParseDuration(v); err == nil {
				fields["serialVol.timeout"] = t
				opts = append(opts, serialvolume.WithTimeout(t))
			}
		}

		// Check for etcd
		if v, _ := csictx.LookupEnv(
			ctx, EnvVarSerialVolAccessEtcdEndpoints); v != "" {

			fields["serialVol.etcd.endpoints"] = v

			if v, _ := csictx.LookupEnv(
				ctx, EnvVarSerialVolAccessEtcdDomain); v != "" {
				fields["serialVol.etcd.domain"] = v
			}

			p, err := etcd.New(ctx, "", etcd.NewConfig())
			if err != nil {
				log.Fatal(err)
			}
			opts = append(opts, serialvolume.WithLockProvider(p))
		}

		sp.Interceptors = append(sp.Interceptors, serialvolume.New(opts...))
		log.WithFields(fields).Debug("enabled serial volume access")
	}

	return
}

func (sp *StoragePlugin) injectContext(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	return handler(csictx.WithLookupEnv(ctx, sp.lookupEnv), req)
}

func (sp *StoragePlugin) getSupportedVersions(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	if info.FullMethod != utils.GetSupportedVersions ||
		len(sp.supportedVersions) == 0 {

		return handler(ctx, req)
	}

	rep := &csi.GetSupportedVersionsResponse{
		SupportedVersions: make([]*csi.Version, len(sp.supportedVersions)),
	}
	for i := range sp.supportedVersions {
		rep.SupportedVersions[i] = &sp.supportedVersions[i]
	}

	return rep, nil
}

func (sp *StoragePlugin) getPluginInfo(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	if info.FullMethod != utils.GetPluginInfo || sp.pluginInfo.Name == "" {
		return handler(ctx, req)
	}
	return &sp.pluginInfo, nil
}
