package csp

const usage = `NAME
    {{.Name}} -- {{.Description}}

SYNOPSIS
    {{.BinPath}}
{{if .Usage}}
STORAGE OPTIONS
{{.Usage}}{{end}}
GLOBAL OPTIONS
    CSI_ENDPOINT
        The CSI endpoint may also be specified by the environment variable
        CSI_ENDPOINT. The endpoint should adhere to Go's network address
        pattern:

            * tcp://host:port
            * unix:///path/to/file.sock.

        If the network type is omitted then the value is assumed to be an
        absolute or relative filesystem path to a UNIX socket file

    X_CSI_DEBUG
        Enabling this option is the same as:
            X_CSI_LOG_LEVEL=debug
            X_CSI_REQ_LOGGING=true
            X_CSI_REP_LOGGING=true

    X_CSI_LOG_LEVEL
        The log level. Valid values include:
           * PANIC
           * FATAL
           * ERROR
           * WARN
           * INFO
           * DEBUG

        The default value is WARN.

    X_CSI_SUPPORTED_VERSIONS
        A space-delimited list of versions formatted MAJOR.MINOR.PATCH.
        Setting this environment variable will cause the program to
        bypass the SP's GetSupportedVersions RPC and return the list of
        specified versions instead.

    X_CSI_PLUGIN_INFO
        The plug-in information is specified via the following
        comma-separated format:

            NAME,VENDOR_VERSION,MANIFEST

        The MANIFEST value may be a series of key/value pairs where either
        the key or value may be quoted to preserve leading or trailing
        whitespace. For example:

            key1=val1 key2="val2 " "key 3"=' val3'

        Setting this environment variable will cause the program to
        bypass the SP's GetPluginInfo RPC and returns the specified
        information instead.

    X_CSI_REQ_LOGGING
        A flag that enables logging of incoming requests to STDOUT.

        Enabling this option sets X_CSI_REQ_ID_INJECTION=true.

    X_CSI_REP_LOGGING
        A flag that enables logging of outgoing responses to STDOUT.

        Enabling this option sets X_CSI_REQ_ID_INJECTION=true.

    X_CSI_REQ_ID_INJECTION
        A flag that enables request ID injection. The ID is parsed from
        the incoming request's metadata with a key of "csi.requestid".
        If no value for that key is found then a new request ID is
        generated using an atomic sequence counter.

    X_CSI_SPEC_VALIDATION
        A flag that enables validation of incoming requests and outgoing
        responses against the CSI specification.

    X_CSI_CREATE_VOL_ALREADY_EXISTS
        A flag that enables treating CreateVolume responses as successful
        when they have an associated error code of AlreadyExists.

        Enabling this option sets X_CSI_SPEC_VALIDATION=true.

    X_CSI_DELETE_VOL_NOT_FOUND
        A flag that enables treating DeleteVolume responses as successful
        when they have an associated error code of NotFound.

        Enabling this option sets X_CSI_SPEC_VALIDATION=true.

    X_CSI_REQUIRE_NODE_ID
        A flag that enables treating the following fields as required:
            * ControllerPublishVolumeRequest.NodeId
            * GetNodeIDResponse.NodeId

        Enabling this option sets X_CSI_SPEC_VALIDATION=true.

    X_CSI_REQUIRE_PUB_VOL_INFO
        A flag that enables treating the following fields as required:
            * ControllerPublishVolumeResponse.PublishVolumeInfo
            * NodePublishVolumeRequest.PublishVolumeInfo

        Enabling this option sets X_CSI_SPEC_VALIDATION=true.

    X_CSI_REQUIRE_VOL_ATTRIBS
        A flag that enables treating the following fields as required:
            * ControllerPublishVolumeRequest.VolumeAttributes
            * ValidateVolumeCapabilitiesRequest.VolumeAttributes
            * NodePublishVolumeRequest.VolumeAttributes

        Enabling this option sets X_CSI_SPEC_VALIDATION=true.

    X_CSI_REQUIRE_CREDS
        Setting X_CSI_REQUIRE_CREDS=true is the same as:
            X_CSI_REQUIRE_CREDS_CREATE_VOL=true
            X_CSI_REQUIRE_CREDS_DELETE_VOL=true
            X_CSI_REQUIRE_CREDS_CTRLR_PUB_VOL=true
            X_CSI_REQUIRE_CREDS_CTRLR_UNPUB_VOL=true
            X_CSI_REQUIRE_CREDS_NODE_PUB_VOL=true
            X_CSI_REQUIRE_CREDS_NODE_UNPUB_VOL=true

        Enabling this option sets X_CSI_SPEC_VALIDATION=true.

    X_CSI_REQUIRE_CREDS_CREATE_VOL
        A flag that enables treating the following fields as required:
            * CreateVolumeRequest.UserCredentials

        Enabling this option sets X_CSI_SPEC_VALIDATION=true.

    X_CSI_REQUIRE_CREDS_DELETE_VOL
        A flag that enables treating the following fields as required:
            * DeleteVolumeRequest.UserCredentials

        Enabling this option sets X_CSI_SPEC_VALIDATION=true.

    X_CSI_REQUIRE_CREDS_CTRLR_PUB_VOL
        A flag that enables treating the following fields as required:
            * ControllerPublishVolumeRequest.UserCredentials

        Enabling this option sets X_CSI_SPEC_VALIDATION=true.

    X_CSI_REQUIRE_CREDS_CTRLR_UNPUB_VOL
        A flag that enables treating the following fields as required:
            * ControllerUnpublishVolumeRequest.UserCredentials

        Enabling this option sets X_CSI_SPEC_VALIDATION=true.

    X_CSI_REQUIRE_CREDS_NODE_PUB_VOL
        A flag that enables treating the following fields as required:
            * NodePublishVolumeRequest.UserCredentials

        Enabling this option sets X_CSI_SPEC_VALIDATION=true.

    X_CSI_REQUIRE_CREDS_NODE_UNPUB_VOL
        A flag that enables treating the following fields as required:
            * NodeUnpublishVolumeRequest.UserCredentials

        Enabling this option sets X_CSI_SPEC_VALIDATION=true.

    X_CSI_IDEMP
        A flag that enables the idempotency interceptor. Even when true,
        the StoragePlugin must still provide a valid IdempotencyProvider
        in order to enable the idempotency interceptor.

    X_CSI_IDEMP_TIMEOUT
        A time.Duration string that determines how long the idempotency
        interceptor waits to obtain a lock for the request's volume before
        returning a the gRPC error code FailedPrecondition (5) to indicate
        an operation is already pending for the specified volume.

    X_CSI_IDEMP_REQUIRE_VOL
        A flag that indicates whether the idempotency interceptor validates
        the existence of a volume before allowing an operation to proceed.

The flags -?,-h,-help may be used to print this screen.
`
