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

    X_CSI_ENDPOINT_PERMS
        When CSI_ENDPOINT is set to a UNIX socket file this environment
        variable may be used to specify the socket's file permissions
        as an octal number, ex. 0644. Please note this value has no
        effect if CSI_ENDPOINT specifies a TCP socket.

        The default value is 0755.

    X_CSI_ENDPOINT_USER
        When CSI_ENDPOINT is set to a UNIX socket file this environment
        variable may be used to specify the UID or user name of the
        user that owns the file. Please note this value has no
        effect if CSI_ENDPOINT specifies a TCP socket.

        If no value is specified then the user owner of the file is the
        same as the user that starts the process.

    X_CSI_ENDPOINT_GROUP
        When CSI_ENDPOINT is set to a UNIX socket file this environment
        variable may be used to specify the GID or group name of the
        group that owns the file. Please note this value has no
        effect if CSI_ENDPOINT specifies a TCP socket.

        If no value is specified then the group owner of the file is the
        same as the group that starts the process.

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
        A list of comma-separated versions strings: MAJOR.MINOR.PATCH.
        Setting this environment variable will cause the program to
        bypass the SP's GetSupportedVersions RPC and return the list of
        specified versions instead.

    X_CSI_PLUGIN_INFO
        The plug-in information is specified via the following
        comma-separated format:

            NAME, VENDOR_VERSION[, MANIFEST...]

        The MANIFEST value may be a series of additional
        comma-separated key/value pairs.

        Please see the encoding/csv package (https://goo.gl/1j1xb9) for
        information on how to quote keys and/or values to include
        leading and trailing whitespace.

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

    X_CSI_PRIVATE_MOUNT_DIR
        Specifies the path of the private mount directory. During a
        NodePublishVolume RPC, the SP will mount a device into the
        private mount area depending on the volume capability:

            * For a Block capability the device will be bind mounted
              to a file in the private mount directory.
            * For a Mount capability the device will be mounted to a
              directory in the private mount directory.

        The SP then bind mounts the private mount to the target path
        specified in the NodePublishVolumeRequest.

The flags -?,-h,-help may be used to print this screen.
`
