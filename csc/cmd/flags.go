package cmd

import (
	"strconv"
	"time"

	flag "github.com/spf13/pflag"
)

// flagEndpoint adds the -e,--endpoint flag to the specified flagset.
func flagEndpoint(fs *flag.FlagSet, addr *string, def string) {
	fs.StringVarP(
		addr,
		"endpoint",
		"e",
		def,
		`The CSI endpoint may also be specified by the environment variable
        CSI_ENDPOINT. The endpoint should adhere to Go's network address
        pattern:

            * tcp://host:port
            * unix:///path/to/file.sock.

        If the network type is omitted then the value is assumed to be an
        absolute or relative filesystem path to a UNIX socket file`)
}

// flagLogLevel adds the -l,--log-level flag to the specified flagset.
func flagLogLevel(fs *flag.FlagSet, addr *logLevelArg, def string) {
	if def != "" {
		addr.Set(def)
	}
	fs.VarP(
		addr,
		"log-level",
		"l",
		`Sets the log level`)
}

// flagTimeout adds the -t,--timeout flag to the specified flagset.
func flagTimeout(fs *flag.FlagSet, addr *time.Duration, def string) {
	t, _ := time.ParseDuration(def)
	fs.DurationVarP(
		addr,
		"timeout",
		"t",
		t,
		`A duration string that specifies the timeout used for gRPC operations.
        Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", and
        "h"`)
}

// flagWithRequestLogging adds the --with-request-logging flag to the
// specified flagset.
func flagWithRequestLogging(fs *flag.FlagSet, addr *bool, def string) {
	fs.BoolVar(
		addr,
		"with-request-logging",
		defBool(def),
		`Enables gRPC request logging. Please note that gRPC requests are
        logged at the INFO log level; so please adjust the log level accordingly.`)
}

// flagWithResponseLogging adds the --with-response-logging flag to the
// specified flagset.
func flagWithResponseLogging(fs *flag.FlagSet, addr *bool, def string) {
	fs.BoolVar(
		addr,
		"with-response-logging",
		defBool(def),
		`Enables gRPC response logging. Please note that gRPC responses are
        logged at the INFO log level; so please adjust the log level accordingly.`)
}

// flagWithSpecValidation adds the --with-spec-validation flag to the
// specified flagset.
func flagWithSpecValidation(fs *flag.FlagSet, addr *bool, def string) {
	fs.BoolVar(
		addr,
		"with-spec-validation",
		defBool(def),
		`Enables validation of outgoing and incoming gRPC requests and responses
        against the CSI specification.`)
}

// flagWithRequiresCreds adds the flag --with-requires-creds
// to the provided flagset.
func flagWithRequiresCreds(fs *flag.FlagSet, addr *bool, def string) {
	fs.BoolVar(
		addr,
		"with-requires-creds",
		defBool(def),
		`Marks a request's UserCredentials field as required.
        Enabling this option also enables --with-spec-validation.`)
}

// flagWithRequiresAttribs adds the flag --with-requires-attribs
// to the provided flagset.
func flagWithRequiresAttribs(fs *flag.FlagSet, addr *bool, def string) {
	fs.BoolVar(
		addr,
		"with-requires-attribs",
		defBool(def),
		`Marks a request's and repsonse's VolumeAttributes field as required.
        Enabling this option also enables --with-spec-validation.`)
}

// flagWithSuccessAlreadyExists adds the flag --with-success-already-exists
// to the provided flagset.
func flagWithSuccessAlreadyExists(fs *flag.FlagSet, addr *bool, def string) {
	fs.BoolVar(
		addr,
		"with-success-create-already-exists",
		defBool(def),
		`Treats a CreateVolume response with an AlreadyExists error
        code as a successful result. Enabling this option also enables
        --with-spec-validation.`)
}

// flagWithSuccessNotFound adds the flag --with-success-not-found
// to the provided flagset.
func flagWithSuccessNotFound(fs *flag.FlagSet, addr *bool, def string) {
	fs.BoolVar(
		addr,
		"with-success-delete-not-found",
		defBool(def),
		`Treats a DeleteVolume response with a NotFound error code
        as a successful result. Enabling this option also enables
        --with-spec-validation.`)
}

func defBool(def string) bool {
	if def == "" {
		return false
	}
	b, _ := strconv.ParseBool(def)
	return b
}

// flagVolumeCapability adds the --cap flag to the specified flagset.
func flagVolumeCapability(fs *flag.FlagSet, addr *volumeCapabilitySliceArg) {
	fs.Var(
		addr,
		"cap",
		`The volume capability is specified using the following format:
`+flagVolumeCapabilityDescSuffix)
}

// flagVolumeAttributes adds the --attrib flag to the specified flagset.
func flagVolumeAttributes(fs *flag.FlagSet, addr *mapOfStringArg) {
	fs.Var(
		addr,
		"attrib",
		`One or more key/value pairs may be specified to send with
        the request as its VolumeAttributes field:

            --attrib key1=val1,key2=val2 --attrib=key3=val3`)
}

// flagVolumeCapabilities adds the --cap flag to the specified flagset.
func flagVolumeCapabilities(fs *flag.FlagSet, addr *volumeCapabilitySliceArg) {
	fs.Var(
		addr,
		"cap",
		`One or more volume capabilities may be specified using the following
        format:
`+flagVolumeCapabilityDescSuffix)
}

const flagVolumeCapabilityDescSuffix = `
            ACCESS_MODE,ACCESS_TYPE[,FS_TYPE,MOUNT_FLAGS]

        The ACCESS_MODE and ACCESS_TYPE values are required. Their values
        may be the their string name or their gRPC integer value. For example,
        the following two options are equivalent:

            --cap 5,1
            --cap MULTI_NODE_MULTI_WRITER,block

        If the access type specified is "mount" (or its gRPC field value of 2)
        then it's possible to specify a filesystem type and mount flags for
        the volume capability. Multiple mount flags may be specified using
        commas. For example:

            --cap MULTI_NODE_MULTI_WRITER,mount,xfs,uid=500,gid=500`
