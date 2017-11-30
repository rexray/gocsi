package csp

import (
	"context"
	"strconv"
	"strings"

	"github.com/thecodeteam/gocsi"
)

const (
	// EnvVarEndpoint is the name of the environment variable used to
	// specify the CSI endpoint.
	EnvVarEndpoint = "CSI_ENDPOINT"

	// EnvVarEndpointPerms is the name of the environment variable used
	// to specify the file permissions for the CSI endpoint when it is
	// a UNIX socket file. This setting has no effect if CSI_ENDPOINT
	// specifies a TCP socket. The default value is 0755.
	EnvVarEndpointPerms = "X_CSI_ENDPOINT_PERMS"

	// EnvVarEndpointUser is the name of the environment variable used
	// to specify the UID or name of the user that owns the endpoint's
	// UNIX socket file. This setting has no effect if CSI_ENDPOINT
	// specifies a TCP socket. The default value is the user that starts
	// the process.
	EnvVarEndpointUser = "X_CSI_ENDPOINT_USER"

	// EnvVarEndpointGroup is the name of the environment variable used
	// to specify the GID or name of the group that owns the endpoint's
	// UNIX socket file. This setting has no effect if CSI_ENDPOINT
	// specifies a TCP socket. The default value is the group that starts
	// the process.
	EnvVarEndpointGroup = "X_CSI_ENDPOINT_GROUP"

	// EnvVarDebug is the name of the environment variable used to
	// determine whether or not debug mode is enabled.
	//
	// Setting this environment variable to a truthy value is the
	// equivalent of X_CSI_LOG_LEVEL=DEBUG, X_CSI_REQ_LOGGING=true,
	// and X_CSI_REP_LOGGING=true.
	EnvVarDebug = "X_CSI_DEBUG"

	// EnvVarLogLevel is the name of the environment variable used to
	// specify the log level. Valid values include PANIC, FATAL, ERROR,
	// WARN, INFO, and DEBUG.
	EnvVarLogLevel = "X_CSI_LOG_LEVEL"

	// EnvVarSupportedVersions is the name of the environment variable used
	// to specify a list of comma-separated versions supported by the SP. If
	// no value is specified then the SP does not perform a version check on
	// the RPC.
	EnvVarSupportedVersions = "X_CSI_SUPPORTED_VERSIONS"

	// EnvVarPluginInfo is the name of the environment variable used to
	// specify the plug-in info in the format:
	//
	//         NAME, VENDOR_VERSION[, MANIFEST...]
	//
	// The MANIFEST value may be a series of additional comma-separated
	// key/value pairs.
	//
	// Please see the encoding/csv package (https://goo.gl/1j1xb9) for
	// information on how to quote keys and/or values to include leading
	// and trailing whitespace.
	//
	// Setting this environment variable will cause the program to
	// bypass the SP's GetPluginInfo RPC and returns the specified
	// information instead.
	EnvVarPluginInfo = "X_CSI_PLUGIN_INFO"

	// EnvVarNodeSvcOnly is the name of the environment variable
	// used to specify that only the CSI Node Service should be started,
	// meaning that the Controller service should not
	EnvVarNodeSvcOnly = "X_CSI_NODESVC_ONLY"

	// EnvVarCtrlSvcOnly is the name of the environment variable
	// used to specify that only the CSI Controller Service should be
	// started, meaning that the Node service should not
	EnvVarCtrlSvcOnly = "X_CSI_CTRLSVC_ONLY"

	// EnvVarReqLogging is the name of the environment variable
	// used to determine whether or not to enable request logging.
	//
	// Setting this environment variable to a truthy value enables
	// request logging to STDOUT.
	EnvVarReqLogging = "X_CSI_REQ_LOGGING"

	// EnvVarRepLogging is the name of the environment variable
	// used to determine whether or not to enable response logging.
	//
	// Setting this environment variable to a truthy value enables
	// response logging to STDOUT.
	EnvVarRepLogging = "X_CSI_REP_LOGGING"

	// EnvVarReqIDInjection is the name of the environment variable
	// used to determine whether or not to enable request ID injection.
	EnvVarReqIDInjection = "X_CSI_REQ_ID_INJECTION"

	// EnvVarSpecValidation is the name of the environment variable
	// used to determine whether or not to enable validation of incoming
	// requests and outgoing responses against the CSI specification.
	EnvVarSpecValidation = "X_CSI_SPEC_VALIDATION"

	// EnvVarCreateVolAlreadyExistsSuccess is the name of the environment
	// variable used to determine whether or not to treat CreateVolume
	// responses with an error code of AlreadyExists as a successful.
	EnvVarCreateVolAlreadyExistsSuccess = "X_CSI_CREATE_VOL_ALREADY_EXISTS"

	// EnvVarDeleteVolNotFoundSuccess is the name of the environment
	// variable used to determine whether or not to treat DeleteVolume
	// responses with an error code of NotFound as a successful.
	EnvVarDeleteVolNotFoundSuccess = "X_CSI_DELETE_VOL_NOT_FOUND"

	// EnvVarRequireNodeID is the name of the environment variable used
	// to determine whether or not the node ID value is required for
	// requests that accept it and responses that return it such as
	// ControllerPublishVolume and GetNodeId.
	EnvVarRequireNodeID = "X_CSI_REQUIRE_NODE_ID"

	// EnvVarRequirePubVolInfo is the name of the environment variable used
	// to determine whether or not publish volume info is required for
	// requests that accept it and responses that return it such as
	// NodePublishVolume and ControllerPublishVolume.
	EnvVarRequirePubVolInfo = "X_CSI_REQUIRE_PUB_VOL_INFO"

	// EnvVarRequireVolAttribs is the name of the environment variable used
	// to determine whether or not volume attributes are required for
	// requests that accept them and responses that return them such as
	// ControllerPublishVolume and CreateVolume.
	EnvVarRequireVolAttribs = "X_CSI_REQUIRE_VOL_ATTRIBS"

	// EnvVarCreds is the name of the environment variable
	// used to determine whether or not user credentials are required for
	// all RPCs. This value may be overridden for specific RPCs.
	EnvVarCreds = "X_CSI_REQUIRE_CREDS"

	// EnvVarCredsCreateVol is the name of the environment variable
	// used to determine whether or not user credentials are required for
	// the eponymous RPC.
	EnvVarCredsCreateVol = "X_CSI_REQUIRE_CREDS_CREATE_VOL"

	// EnvVarCredsDeleteVol is the name of the environment variable
	// used to determine whether or not user credentials are required for
	// the eponymous RPC.
	EnvVarCredsDeleteVol = "X_CSI_REQUIRE_CREDS_DELETE_VOL"

	// EnvVarCredsCtrlrPubVol is the name of the environment
	// variable used to determine whether or not user credentials are required
	// for the eponymous RPC.
	EnvVarCredsCtrlrPubVol = "X_CSI_REQUIRE_CREDS_CTRLR_PUB_VOL"

	// EnvVarCredsCtrlrUnpubVol is the name of the
	// environment variable used to determine whether or not user credentials
	// are required for the eponymous RPC.
	EnvVarCredsCtrlrUnpubVol = "X_CSI_REQUIRE_CREDS_CTRLR_UNPUB_VOL"

	// EnvVarCredsNodePubVol is the name of the environment
	// variable used to determine whether or not user credentials are required
	// for the eponymous RPC.
	EnvVarCredsNodePubVol = "X_CSI_REQUIRE_CREDS_NODE_PUB_VOL"

	// EnvVarCredsNodeUnpubVol is the name of the environment
	// variable used to determine whether or not user credentials are required
	// for the eponymous RPC.
	EnvVarCredsNodeUnpubVol = "X_CSI_REQUIRE_CREDS_NODE_UNPUB_VOL"

	// EnvVarIdemp is the name of the environment variable
	// used to determine whether or not to enable idempotency.
	EnvVarIdemp = "X_CSI_IDEMP"

	// EnvVarIdempTimeout is the name of the environment variable
	// used to specify the timeout for gRPC operations.
	EnvVarIdempTimeout = "X_CSI_IDEMP_TIMEOUT"

	// EnvVarIdempRequireVolume is the name of the environment variable
	// used to determine whether or not the idempotency interceptor
	// checks to see if a volume exists before allowing an operation.
	EnvVarIdempRequireVolume = "X_CSI_IDEMP_REQUIRE_VOL"

	// EnvVarPrivateMountDir is the name of the environment variable
	// that specifies the path of the private mount directory used by
	// SPs to mount a device during a NodePublishVolume RPC before
	// bind mounting the file/directory from the private mount area
	// to the target path.
	EnvVarPrivateMountDir = "X_CSI_PRIVATE_MOUNT_DIR"
)

func (sp *StoragePlugin) initEnvVars(ctx context.Context) {

	// Copy the environment variables from the public EnvVar
	// string slice to the private envVars map for quick lookup.
	sp.envVars = map[string]string{}
	for _, v := range sp.EnvVars {
		// Environment variables must adhere to one of the following
		// formats:
		//
		//     - ENV_VAR_KEY=
		//     - ENV_VAR_KEY=ENV_VAR_VAL
		pair := strings.SplitN(v, "=", 2)
		if len(pair) < 1 || len(pair) > 2 {
			continue
		}

		// Ensure the environment variable is stored in all upper-case
		// to make subsequent map-lookups deterministic.
		key := strings.ToUpper(pair[0])

		// Check to see if the value for the key is available from the
		// context's os.Environ or os.LookupEnv functions. If neither
		// return a value then use the provided default value.
		var val string
		if v, ok := gocsi.LookupEnv(ctx, key); ok {
			val = v
		} else if len(pair) > 1 {
			val = pair[1]
		}
		sp.envVars[key] = val
	}

	// Check for the debug value.
	if v, ok := gocsi.LookupEnv(ctx, EnvVarDebug); ok {
		if ok, _ := strconv.ParseBool(v); ok {
			gocsi.Setenv(ctx, EnvVarReqLogging, "true")
			gocsi.Setenv(ctx, EnvVarRepLogging, "true")
		}
	}

	return
}

func (sp *StoragePlugin) initSupportedVersions(ctx context.Context) {
	szVersions, ok := gocsi.LookupEnv(ctx, EnvVarSupportedVersions)
	if !ok {
		return
	}
	sp.supportedVersions = gocsi.ParseVersions(szVersions)
}

func (sp *StoragePlugin) initPluginInfo(ctx context.Context) {
	szInfo, ok := gocsi.LookupEnv(ctx, EnvVarPluginInfo)
	if !ok {
		return
	}
	info := strings.SplitN(szInfo, ",", 3)
	if len(info) > 0 {
		sp.pluginInfo.Name = info[0]
	}
	if len(info) > 1 {
		sp.pluginInfo.VendorVersion = info[1]
	}
	if len(info) > 2 {
		sp.pluginInfo.Manifest = gocsi.ParseMap(info[2])
	}
}
