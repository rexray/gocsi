package gocsi

import (
	"bytes"
	"fmt"
)

const usageTemplate = `NAME
    {{.Name}} -- {{.Description}}

SYNOPSIS
    {{.BinPath}}
{{if .AppUsage}}
APP OPTIONS
{{.AppUsage}}{{end}}
GLOBAL OPTIONS
    CSI_ENDPOINT
        The CSI endpoint may also be specified by the environment variable
        CSI_ENDPOINT. The endpoint should adhere to Go's network address
        pattern:

            * tcp://host:port
            * unix:///path/to/file.sock.

        If the network type is omitted then the value is assumed to be an
        absolute or relative filesystem path to a UNIX socket file

    X_CSI_MODE
        Specifies the service mode of the storage plug-in. Valid values are:

            * <empty>
            * controller
            * node

        If unset or set to an empty value the storage plug-in activates
        both controller and node services. The identity service is always
        activated.

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

        If the GoCSI logging middleware is used by the storage plug-in,
        this option enables request and response logging.

    X_CSI_LOG_LEVEL
        The log level. Valid values include:
           * PANIC
           * FATAL
           * ERROR
           * WARN
           * INFO
           * DEBUG

        The default value is WARN.{{if .StoragePluginUsage}}

{{.StoragePluginUsage}}{{end}}
The flags -?,-h,-help may be used to print this screen.
`

type hasUsage interface {
	Usage() string
}

// Usage returns the Storage Plugin's usage string.
func (sp *StoragePlugin) Usage() string {

	w := &bytes.Buffer{}

	// Range over and initialize the provided middleware. For
	// middleware that initializes successfully, add its HandleServer
	// function to the list of gRPC interceptors used by the server.
	for i, j := range sp.Middleware {
		if o, ok := j.(hasUsage); ok {
			if u := o.Usage(); len(u) > 0 {
				fmt.Fprintln(w, u)
				if i < len(sp.Middleware)-1 {
					fmt.Fprintln(w)
				}
			}
		}
	}

	return w.String()
}
