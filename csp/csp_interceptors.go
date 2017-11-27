package csp

import (
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/thecodeteam/gocsi"
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
		withIdemp              = sp.getEnvBool(ctx, EnvVarIdemp)
		withSpec               = sp.getEnvBool(ctx, EnvVarSpecValidation)
		withNewVolExists       = sp.getEnvBool(ctx, envVarNewVolExists)
		withDelVolNotFound     = sp.getEnvBool(ctx, envVarDelVolNotFound)
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
		withNewVolExists ||
		withDelVolNotFound ||
		withNodeID ||
		withPubVolInfo ||
		withVolAttribs

	// Configure logging.
	if withReqLogging || withRepLogging {
		// Automatically enable request ID injection if logging
		// is enabled.
		sp.Interceptors = append(sp.Interceptors,
			gocsi.NewServerRequestIDInjector())
		log.Debug("enabled request ID injector")

		var (
			loggingOpts []gocsi.LoggingOption
			w           = newLogger(log.Infof)
		)

		if withReqLogging {
			loggingOpts = append(loggingOpts, gocsi.WithRequestLogging(w))
			log.Debug("enabled request logging")
		}
		if withRepLogging {
			loggingOpts = append(loggingOpts, gocsi.WithResponseLogging(w))
			log.Debug("enabled response logging")
		}
		sp.Interceptors = append(sp.Interceptors,
			gocsi.NewServerLogger(loggingOpts...))
	}

	if withSpec {
		var specOpts []gocsi.SpecValidatorOption

		if len(sp.supportedVersions) > 0 {
			specOpts = append(
				specOpts,
				gocsi.WithSupportedVersions(sp.supportedVersions...))
		}
		if withCredsNewVol {
			specOpts = append(specOpts,
				gocsi.WithRequiresCreateVolumeCredentials())
			log.Debug("enabled spec validator opt: requires creds: " +
				"CreateVolume")
		}
		if withCredsDelVol {
			specOpts = append(specOpts,
				gocsi.WithRequiresDeleteVolumeCredentials())
			log.Debug("enabled spec validator opt: requires creds: " +
				"DeleteVolume")
		}
		if withCredsCtrlrPubVol {
			specOpts = append(specOpts,
				gocsi.WithRequiresControllerPublishVolumeCredentials())
			log.Debug("enabled spec validator opt: requires creds: " +
				"ControllerPublishVolume")
		}
		if withCredsCtrlrUnpubVol {
			specOpts = append(specOpts,
				gocsi.WithRequiresControllerUnpublishVolumeCredentials())
			log.Debug("enabled spec validator opt: requires creds: " +
				"ControllerUnpublishVolume")
		}
		if withCredsNodePubVol {
			specOpts = append(specOpts,
				gocsi.WithRequiresNodePublishVolumeCredentials())
			log.Debug("enabled spec validator opt: requires creds: " +
				"NodePublishVolume")
		}
		if withCredsNodeUnpubVol {
			specOpts = append(specOpts,
				gocsi.WithRequiresNodeUnpublishVolumeCredentials())
			log.Debug("enabled spec validator opt: requires creds: " +
				"NodeUnpublishVolume")
		}

		if withNodeID {
			specOpts = append(specOpts,
				gocsi.WithRequiresNodeID())
			log.Debug("enabled spec validator opt: requires node ID")
		}
		if withPubVolInfo {
			specOpts = append(specOpts,
				gocsi.WithRequiresPublishVolumeInfo())
			log.Debug("enabled spec validator opt: requires pub vol info")
		}
		if withVolAttribs {
			specOpts = append(specOpts,
				gocsi.WithRequiresVolumeAttributes())
			log.Debug("enabled spec validator opt: requires vol attribs")
		}
		if withNewVolExists {
			specOpts = append(specOpts,
				gocsi.WithSuccessCreateVolumeAlreadyExists())
			log.Debug("enabled spec validator opt: create exists success")
		}
		if withDelVolNotFound {
			specOpts = append(specOpts,
				gocsi.WithSuccessDeleteVolumeNotFound())
			log.Debug("enabled spec validator opt: delete !exists success")
		}
		sp.Interceptors = append(sp.Interceptors,
			gocsi.NewServerSpecValidator(specOpts...))
	}

	if _, ok := gocsi.LookupEnv(ctx, EnvVarPluginInfo); ok {
		log.Debug("enabled GetPluginInfo interceptor")
		sp.Interceptors = append(sp.Interceptors, sp.getPluginInfo)
	}

	if len(sp.supportedVersions) > 0 {
		log.Debug("enabled GetSupportedVersions interceptor")
		sp.Interceptors = append(sp.Interceptors, sp.getSupportedVersions)
	}

	if withIdemp && sp.IdempotencyProvider != nil {
		var (
			opts   []gocsi.IdempotentInterceptorOption
			fields = map[string]interface{}{}
		)

		// Get idempotency provider's timeout.
		if v, _ := gocsi.LookupEnv(ctx, EnvVarIdempTimeout); v != "" {
			if t, err := time.ParseDuration(v); err == nil {
				fields["idemp.timeout"] = t
				opts = append(opts, gocsi.WithIdempTimeout(t))
			}
		}

		// Check to see if the idempotency provider requires volumes to exist.
		if sp.getEnvBool(ctx, EnvVarIdempRequireVolume) {
			fields["idemp.volRequired"] = true
			opts = append(opts, gocsi.WithIdempRequireVolumeExists())
		}

		sp.Interceptors = append(sp.Interceptors,
			gocsi.NewIdempotentInterceptor(sp.IdempotencyProvider, opts...))
		log.WithFields(fields).Debug("enabled idempotency provider")
	}

	return
}

func (sp *StoragePlugin) injectContext(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	return handler(gocsi.WithLookupEnv(ctx, sp.lookupEnv), req)
}

func (sp *StoragePlugin) getSupportedVersions(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	if info.FullMethod != gocsi.GetSupportedVersions ||
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

	if info.FullMethod != gocsi.GetPluginInfo || sp.pluginInfo.Name == "" {
		return handler(ctx, req)
	}
	return &sp.pluginInfo, nil
}
