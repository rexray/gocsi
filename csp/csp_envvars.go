package csp

const (
	// EnvVarEndpoint is the name of the environment variable used to
	// specify the CSI endpoint.
	EnvVarEndpoint = "CSI_ENDPOINT"

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
	// to specify a space-delimited list of versions supported by the SP. If
	// no value is specified then the SP does not perform a version check on
	// the RPC.
	EnvVarSupportedVersions = "X_CSI_SUPPORTED_VERSIONS"

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
)
