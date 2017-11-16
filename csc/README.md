# Container Storage Client
The Container Storage Client (`csc`) is a command line interface (CLI) tool
that provides analogues for all of the CSI RPCs.

```bash
$ csc
a command line client for csi storage plug-ins

Usage:
  csc [command]

Examples:

CSI ENDPOINT

The CSI endpoint is specified with either the environment variable
CSI_ENDPOINT or the flag -e, --endpoint. The specified endpoint value
should adhere to the Go network address pattern(s):

    csc --endpoint tcp://host:port

    csc --endpoint unix://path/to/file.sock

Additionally, if the network type is omitted then this program
assumes the provided endpoint value is the relative or absolute path
to a UNIX socket file:

    csc --endpoint file.sock


USER CREDENTIALS

While this program does support CSI user credentials, there is
no flag for specifying them on the command line. This is a design
choice in order to prevent sensitive information from being part of
a process listing.

User credentials may be specified via the environment variable
X_CSI_USER_CREDENTIALS. The format of this variable supports multiple
credential pairs:

    X_CSI_USER_CREDENTIALS=user1=pass user2="pass with trailing space "

As illustrated above, the value of the enviroment variable is one
or more key/value pairs. Both the key and value may be quoted to
preserve whitespace.


VOLUME CAPABILITIES

When specifying volume capabilities on the command line, the following
format is used:

    ACCESS_MODE,ACCESS_TYPE[,FS_TYPE,MOUNT_FLAGS]

The ACCESS_MODE value may be the mode's full name or its integer value.
For example, the following two values are equivalent:

    MULTI_NODE_MULTI_WRITER
    5

The ACCESS_TYPE value may also reflect the type name or numeric value.
For example:

    block
    1

If the ACCESS_TYPE specifies is "mount" (or its numeric equivalent of 2)
then it's also possible to specify a filesystem type and mount flags
for the mount capability. Here are some examples:

    --cap 1,block
    --cap MULTI_NODE_MULTI_WRITER,mount,xfs,uid=500,gid=500


LOGGING

The log level may be adjusted with the flag -l,--log-level. In order to
enable gRPC request or response logging the flags --with-request-logging,
--with-response-logging must also be used. These flags enable the
GoCSI client-side logging interceptor. Please note that this interceptor
logs request and response data at the INFO level, so set the log level
accordingly.


SPEC VALIDATION

Please note that there are many flags, --with-ABC, that enable
client-side request and response validation against the CSI
specification. These flags enable a GoCSI gRPC interceptor to provide
validation. There are also flags that enable optional components of the
spec validation, such as treating the node ID as required, or treating
an ALREADY_EXISTS error from CreateVolume as successful. None of these
options are enabled by default.


Available Commands:
  controller  the csi controller service rpcs
  help        Help about any command
  identity    the csi identity service rpcs
  node        the csi node service rpcs

Flags:
  -e, --endpoint string                  the csi endpoint
  -h, --help                             help for csc
  -i, --insecure                         a flag that disables tls (default true)
  -l, --log-level string                 the log level (default "warn")
  -m, --metadata key=val[,key=val,...]   one or more key/value pairs used as grpc metadata
  -t, --timeout duration                 the timeout used for dialing the csi endpoint and invoking rpcs (default 1m0s)
  -v, --version major.minor.patch        the csi version to send with an rpc
      --with-request-logging             enables request logging
      --with-response-logging            enables response logging
      --with-spec-validation             enables validation of request/response data against the CSI specification

Use "csc [command] --help" for more information about a command.
```
