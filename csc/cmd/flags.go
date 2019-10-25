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

// flagWithRequiresVolContext adds the flag --with-requires-vol-context
// to the provided flagset
func flagWithRequiresVolContext(fs *flag.FlagSet, addr *bool, def bool) {
	fs.BoolVar(
		addr,
		"with-requires-vol-context",
		def,
		`Marks a request's and repsonse's VolumeContext field as required.
        Enabling this option also enables --with-spec-validation.`)
}

// flagWithRequiresPubContext adds the flag --with-requires-pub-context
// to the provided flagset
func flagWithRequiresPubContext(fs *flag.FlagSet, addr *bool, def bool) {
	fs.BoolVar(
		addr,
		"with-requires-pub-context",
		def,
		`Marks a request's and repsonse's PublishContext field as required.
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

// flagVolumeContext adds the --vol-context flag to the specified flagset.
func flagVolumeContext(fs *flag.FlagSet, addr *mapOfStringArg) {
	fs.Var(
		addr,
		"vol-context",
		`One or more key/value pairs may be specified to send with
        the request as its VolumeContext field:

            --vol-context key1=val1,key2=val2 --vol-context=key3=val3`)
}

// flagPublishContext adds the --pub-context flag to the specified flagset.
func flagPublishContext(fs *flag.FlagSet, addr *mapOfStringArg) {
	fs.Var(
		addr,
		"pub-context",
		`One or more key/value pairs may be specified to send with
        the request as its PublishContext field:

            --pub-context key1=val1,key2=val2 --pub-context=key3=val3`)
}

// flagParameters adds the --params flag to the specified flagset.
func flagParameters(fs *flag.FlagSet, addr *mapOfStringArg) {
	fs.Var(
		addr,
		"params",
		`One or more key/value pairs may be specified to send with
        the request as its Parameters field:

            --params key1=val1,key2=val2 --params=key3=val3`)
}

// flagStagingTargetPath adds the --staging-target-path flag to the specified
// flagset.
func flagStagingTargetPath(fs *flag.FlagSet, addr *string) {
	fs.StringVar(
		addr,
		"staging-target-path",
		"",
		"The path to which to stage or unstage the volume")
}

// flagTargetPath adds the --target-path flag to the specified flagset.
func flagTargetPath(fs *flag.FlagSet, addr *string) {
	fs.StringVar(
		addr,
		"target-path",
		"",
		"The path to which to mount or unmount the volume")
}

// flagVolumeSrc adds the --from-volume flag to specified flagset
func flagVolumeSrc(fs *flag.FlagSet, vname *string) {
	fs.StringVar(
		vname,
		"volume-src",
		"",
		"The name of the source volume")
}

// flagSnapshotSrc adds the --from-snapshot flag to specified flagset
func flagSnapshotSrc(fs *flag.FlagSet, sname *string) {
	fs.StringVar(
		sname,
		"snapshot-src",
		"",
		"The name of the source snapshot")
}

// flagReadOnly adds the --read-only flag to the specified flagset
func flagReadOnly(fs *flag.FlagSet, addr *bool) {
	fs.BoolVar(
		addr,
		"read-only",
		false,
		"Mark the volume as read-only")
}

// flagRequiredBytes adds the --req-bytes flag to the specified flagset
func flagRequiredBytes(fs *flag.FlagSet, addr *int64) {
	fs.Int64Var(
		addr,
		"req-bytes",
		0,
		"The required size of the volume in bytes")
}

// flagLimitBytes adds the --lim-bytes flag to the specified flagset
func flagLimitBytes(fs *flag.FlagSet, addr *int64) {
	fs.Int64Var(
		addr,
		"lim-bytes",
		0,
		"The limit to the size of the volume in bytes")
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
