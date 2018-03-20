package env

import (
	"context"
	"os"
	"strconv"
	"strings"
)

const (
	// Endpoint is the name of the environment variable used to
	// specify the CSI endpoint.
	Endpoint = "CSI_ENDPOINT"

	// EndpointPerms is the name of the environment variable used
	// to specify the file permissions for the CSI endpoint when it is
	// a UNIX socket file. This setting has no effect if CSI_ENDPOINT
	// specifies a TCP socket. The default value is 0755.
	EndpointPerms = "X_CSI_ENDPOINT_PERMS"

	// EndpointUser is the name of the environment variable used
	// to specify the UID or name of the user that owns the endpoint's
	// UNIX socket file. This setting has no effect if CSI_ENDPOINT
	// specifies a TCP socket. The default value is the user that starts
	// the process.
	EndpointUser = "X_CSI_ENDPOINT_USER"

	// EndpointGroup is the name of the environment variable used
	// to specify the GID or name of the group that owns the endpoint's
	// UNIX socket file. This setting has no effect if CSI_ENDPOINT
	// specifies a TCP socket. The default value is the group that starts
	// the process.
	EndpointGroup = "X_CSI_ENDPOINT_GROUP"

	// Mode is the name of the environment variable used to specify
	// the service mode of the storage plug-in. Valie values are:
	//
	// * <empty>
	// * controller
	// * node
	//
	// If unset or set to an empty value the storage plug-in activates
	// both controller and node services. The identity service is always
	// activated.
	Mode = "X_CSI_MODE"

	// Debug is the name of the environment variable used to
	// determine whether or not debug mode is enabled.
	Debug = "X_CSI_DEBUG"

	// LogLevel is the name of the environment variable used to
	// specify the log level. Valid values include PANIC, FATAL, ERROR,
	// WARN, INFO, and DEBUG.
	LogLevel = "X_CSI_LOG_LEVEL"
)

////////////////////////////////////////////////////////////////////////////////
///                                  ETCD                                    ///
////////////////////////////////////////////////////////////////////////////////
const (
	// EtcdEndpoints is the name of the environment
	// variable that defines the etcd endoints.
	EtcdEndpoints = "X_CSI_ETCD_ENDPOINTS"

	// EtcdPrefix is the name of the environment
	// variable that defines the etcd prefix.
	EtcdPrefix = "X_CSI_ETCD_PREFIX"

	// EtcdAutoSyncInterval is the name of the environment
	// variable that defines the interval to update endpoints with its latest
	//  members. 0 disables auto-sync. By default auto-sync is disabled.
	EtcdAutoSyncInterval = "X_CSI_ETCD_AUTO_SYNC_INTERVAL"

	// EtcdDialTimeout is the name of the environment
	// variable that defines the timeout for failing to establish a connection.
	EtcdDialTimeout = "X_CSI_ETCD_DIAL_TIMEOUT"

	// EtcdDialKeepAliveTime is the name of the environment
	// variable that defines the time after which client pings the server to see
	// if transport is alive.
	EtcdDialKeepAliveTime = "X_CSI_ETCD_DIAL_KEEP_ALIVE_TIME"

	// EtcdDialKeepAliveTimeout is the name of the
	// environment variable that defines the time that the client waits for a
	// response for the keep-alive probe. If the response is not received in
	// this time, the connection is closed.
	EtcdDialKeepAliveTimeout = "X_CSI_ETCD_DIAL_KEEP_ALIVE_TIMEOUT"

	// EtcdMaxCallSendMsgSz is the name of the environment
	// variable that defines the client-side request send limit in bytes.
	// If 0, it defaults to 2.0 MiB (2 * 1024 * 1024).
	// Make sure that "MaxCallSendMsgSize" < server-side default send/recv
	// limit. ("--max-request-bytes" flag to etcd or
	// "embed.Config.MaxRequestBytes").
	EtcdMaxCallSendMsgSz = "X_CSI_ETCD_MAX_CALL_SEND_MSG_SZ"

	// EtcdMaxCallRecvMsgSz is the name of the environment
	// variable that defines the client-side response receive limit.
	// If 0, it defaults to "math.MaxInt32", because range response can
	// easily exceed request send limits.
	// Make sure that "MaxCallRecvMsgSize" >= server-side default send/recv
	// limit. ("--max-request-bytes" flag to etcd or
	// "embed.Config.MaxRequestBytes").
	EtcdMaxCallRecvMsgSz = "X_CSI_ETCD_MAX_CALL_RECV_MSG_SZ"

	// EtcdUsername is the name of the environment
	// variable that defines the user name used for authentication.
	EtcdUsername = "X_CSI_ETCD_USERNAME"

	// EtcdPassword is the name of the environment
	// variable that defines the password used for authentication.
	EtcdPassword = "X_CSI_ETCD_PASSWORD"

	// EtcdRejectOldCluster is the name of the environment
	// variable that defines when set will refuse to create a client against
	// an outdated cluster.
	EtcdRejectOldCluster = "X_CSI_ETCD_REJECT_OLD_CLUSTER"

	// EtcdTLS is the name of the environment
	// variable that defines whether or not the client should attempt
	// to use TLS when connecting to the server.
	EtcdTLS = "X_CSI_ETCD_TLS"

	// EtcdTLSInsecure is the name of the environment
	// variable that defines whether or not the TLS connection should
	// verify certificates.
	EtcdTLSInsecure = "X_CSI_ETCD_TLS_INSECURE"
)

////////////////////////////////////////////////////////////////////////////////
///                                 UTILS                                    ///
////////////////////////////////////////////////////////////////////////////////
var (
	// ctxOSEnviron is an interface-wrapped key used to access a string
	// slice that contains one or more environment variables stored as
	// KEY=VALUE.
	ctxOSEnviron = interface{}("os.Environ")

	// ctxOSLookupEnvKey is an interface-wrapped key used to access a function
	// with the signature func(string) (string, bool) that returns the value of
	// an environment variable.
	ctxOSLookupEnvKey = interface{}("os.LookupEnv")

	// ctxOSSetenvKey is an interface-wrapped key used to access a function
	// with the signature func(string, string) that can be used to set the
	// value of an environment variable
	ctxOSSetenvKey = interface{}("os.Setenev")
)

type lookupEnvFunc func(string) (string, bool)
type setenvFunc func(string, string) error

// WithEnviron returns a new Context with the provided environment variable
// string slice.
func WithEnviron(ctx context.Context, v []string) context.Context {
	return context.WithValue(ctx, ctxOSEnviron, v)
}

// GetEnviron returns the environment variable string slice if present in
// the context.
func GetEnviron(ctx context.Context) ([]string, bool) {
	v, ok := ctx.Value(ctxOSEnviron).([]string)
	return v, ok
}

// WithLookupEnv returns a new Context with the provided function.
func WithLookupEnv(ctx context.Context, f lookupEnvFunc) context.Context {
	return context.WithValue(ctx, ctxOSLookupEnvKey, f)
}

// WithSetenv returns a new Context with the provided function.
func WithSetenv(ctx context.Context, f setenvFunc) context.Context {
	return context.WithValue(ctx, ctxOSSetenvKey, f)
}

// LookupEnv returns the value of the provided environment variable by:
//
//   1. Inspecting the context for a key "os.Environ" with a string
//      slice value. If such a key and value exist then the string slice
//      is searched for the specified key and if found its value is returned.
//
//   2. Inspecting the context for a key "os.LookupEnv" with a value of
//      func(string) (string, bool). If such a key and value exist then the
//      function is used to attempt to discover the key's value. If the
//      key and value are found they are returned.
//
//   3. Returning the result of os.LookupEnv.
func LookupEnv(ctx context.Context, key string) (string, bool) {
	if s, ok := ctx.Value(ctxOSEnviron).([]string); ok {
		for _, v := range s {
			p := strings.SplitN(v, "=", 2)
			if len(p) > 0 && strings.EqualFold(p[0], key) {
				if len(p) > 1 {
					return p[1], true
				}
				return "", true
			}
		}
	}
	if f, ok := ctx.Value(ctxOSLookupEnvKey).(lookupEnvFunc); ok {
		if v, ok := f(key); ok {
			return v, true
		}
	}
	return os.LookupEnv(key)
}

// Getenv is an alias for LookupEnv and drops the boolean return value.
func Getenv(ctx context.Context, key string) string {
	val, _ := LookupEnv(ctx, key)
	return val
}

// Setenv sets the value of the provided environment variable to the
// specified value by first inspecting the context for a key "os.Setenv"
// with a value of func(string, string) error. If the context does not
// contain such a function then os.Setenv is used instead.
func Setenv(ctx context.Context, key, val string) error {
	if f, ok := ctx.Value(ctxOSSetenvKey).(setenvFunc); ok {
		return f(key, val)
	}
	return os.Setenv(key, val)
}

// IsDebug indicats whether X_CSI_DEBUG is set to a truthy value.
func IsDebug(ctx context.Context) bool {
	if v, ok := LookupEnv(ctx, Debug); ok {
		b, _ := strconv.ParseBool(v)
		return b
	}
	return false
}
