package gocsi

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"sync/atomic"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/codedellemc/gocsi/csi"
)

// ChainUnaryServerOpt chains one or more unary, server interceptors
// and returns a server option.
func ChainUnaryServerOpt(i ...grpc.UnaryServerInterceptor) grpc.ServerOption {
	return grpc.UnaryInterceptor(ChainUnaryServer(i...))
}

// ChainUnaryServer chains one or more unary, server interceptors
// together into a left-to-right series that can be provided to a
// new gRPC server.
func ChainUnaryServer(
	i ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {

	switch len(i) {
	case 0:
		return func(
			ctx context.Context,
			req interface{},
			_ *grpc.UnaryServerInfo,
			handler grpc.UnaryHandler) (interface{}, error) {
			return handler(ctx, req)
		}
	case 1:
		return i[0]
	}

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {

		bc := func(
			cur grpc.UnaryServerInterceptor,
			nxt grpc.UnaryHandler) grpc.UnaryHandler {
			return func(
				curCtx context.Context,
				curReq interface{}) (interface{}, error) {
				return cur(curCtx, curReq, info, nxt)
			}
		}
		c := handler
		for j := len(i) - 1; j >= 0; j-- {
			c = bc(i[j], c)
		}
		return c(ctx, req)
	}
}

type hasGetVersion interface {
	GetVersion() *csi.Version
}

// VersionValidator validates an incoming message's version against
// the plug-in's supported versions.
type VersionValidator struct {
	// SupportedVersions is a list of the versions supported by the
	// plug-in.
	SupportedVersions []Version
}

// Handle is a unary, server interceptor function.
func (v *VersionValidator) Handle(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	treq, ok := req.(hasGetVersion)
	if !ok {
		return handler(ctx, req)
	}

	rv := treq.GetVersion()

	for _, sv := range v.SupportedVersions {
		if CompareVersions(rv, sv) == 0 {
			return handler(ctx, req)
		}
	}

	msg := fmt.Sprintf(
		"unsupported request version: %s", SprintfVersion(rv))

	switch info.FullMethod {
	case FMControllerGetCapabilities:
		// UNSUPPORTED_REQUEST_VERSION
		return ErrControllerGetCapabilities(2, msg), nil
	case FMControllerPublishVolume:
		// UNSUPPORTED_REQUEST_VERSION
		return ErrControllerPublishVolumeGeneral(2, msg), nil
	case FMControllerUnpublishVolume:
		// UNSUPPORTED_REQUEST_VERSION
		return ErrControllerUnpublishVolumeGeneral(2, msg), nil
	case FMCreateVolume:
		// UNSUPPORTED_REQUEST_VERSION
		return ErrCreateVolumeGeneral(2, msg), nil
	case FMDeleteVolume:
		// UNSUPPORTED_REQUEST_VERSION
		return ErrDeleteVolumeGeneral(2, msg), nil
	case FMGetCapacity:
		// UNSUPPORTED_REQUEST_VERSION
		return ErrGetCapacity(2, msg), nil
	case FMGetNodeID:
		// UNSUPPORTED_REQUEST_VERSION
		return ErrGetNodeIDGeneral(2, msg), nil
	case FMGetPluginInfo:
		// UNSUPPORTED_REQUEST_VERSION
		return ErrGetPluginInfo(2, msg), nil
	case FMListVolumes:
		// UNSUPPORTED_REQUEST_VERSION
		return ErrListVolumes(2, msg), nil
	case FMGetSupportedVersions:
		panic("Version Check Unsupported for GetSupportedVersions")
	case FMNodeGetCapabilities:
		// UNSUPPORTED_REQUEST_VERSION
		return ErrNodeGetCapabilities(2, msg), nil
	case FMNodePublishVolume:
		// UNSUPPORTED_REQUEST_VERSION
		return ErrNodePublishVolumeGeneral(2, msg), nil
	case FMNodeUnpublishVolume:
		// UNSUPPORTED_REQUEST_VERSION
		return ErrNodeUnpublishVolumeGeneral(2, msg), nil
	case FMProbeNode:
		// UNSUPPORTED_REQUEST_VERSION
		return ErrProbeNodeGeneral(2, msg), nil
	case FMValidateVolumeCapabilities:
		// UNSUPPORTED_REQUEST_VERSION
		return ErrValidateVolumeCapabilitiesGeneral(2, msg), nil
	}

	panic("Version Check Unsupported")
}

var requestIDVal uint64

// RequestIDInjector injects a unique request ID into the request.
func RequestIDInjector(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	return handler(
		context.WithValue(
			ctx,
			requestIDKey,
			atomic.AddUint64(&requestIDVal, 1)),
		req)
}

// ServerSideMessageLogger is unary, server interceptor that logs
// incoming messages and the values of any public fields.
type ServerSideMessageLogger struct {
	// Log is the function used to log the message.
	Log func(msg string, args ...interface{})
}

var emptyValRX = regexp.MustCompile(
	`^((?:)|(?:\[\])|(?:<nil>)|(?:map\[\]))$`)

// Handle is a unary, server interceptor function.
func (v *ServerSideMessageLogger) Handle(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	if req == nil {
		return handler(ctx, req)
	}

	w := &bytes.Buffer{}

	if rid, ok := GetRequestID(ctx); ok {
		fmt.Fprintf(w, "%04d: ", rid)
	}

	fmt.Fprintf(w, "%s", info.FullMethod)

	rv := reflect.ValueOf(req).Elem()
	tv := rv.Type()
	nf := tv.NumField()
	printedColon := false
	printComma := false
	for i := 0; i < nf; i++ {
		sv := fmt.Sprintf("%v", rv.Field(i).Interface())
		if emptyValRX.MatchString(sv) {
			continue
		}
		if printComma {
			fmt.Fprintf(w, ", ")
		}
		if !printedColon {
			fmt.Fprintf(w, ": ")
			printedColon = true
		}
		printComma = true
		fmt.Fprintf(w, "%s=%s", tv.Field(i).Name, sv)
	}
	fmt.Fprintln(w)

	v.Log(w.String())
	return handler(ctx, req)
}

// ServerSideInputValidatorFunc is a function that validates an incoming
// server-side RPC message's arguments.
type ServerSideInputValidatorFunc func(
	ctx context.Context, req interface{}) (interface{}, error)

// ServerSideInputValidator can be used to validate messages received
// by a CSI server.
type ServerSideInputValidator map[string]ServerSideInputValidatorFunc

// Private fields that other files in the package can use to set
// the validator functions for the appropriate methods.
var (
	// Controller
	ssvCreateVolume               ServerSideInputValidatorFunc
	ssvDeleteVolume               ServerSideInputValidatorFunc
	ssvControllerPublishVolume    ServerSideInputValidatorFunc
	ssvControllerUnpublishVolume  ServerSideInputValidatorFunc
	ssvValidateVolumeCapabilities ServerSideInputValidatorFunc
	ssvListVolumes                ServerSideInputValidatorFunc
	ssvGetCapacity                ServerSideInputValidatorFunc
	ssvControllerGetCapabilities  ServerSideInputValidatorFunc

	// Identity
	ssvGetSupportedVersions ServerSideInputValidatorFunc
	ssvGetPluginInfo        ServerSideInputValidatorFunc

	// Nodess
	ssvNodePublishVolume   ServerSideInputValidatorFunc
	ssvNodeUnpublishVolume ServerSideInputValidatorFunc
	ssvGetNodeID           ServerSideInputValidatorFunc
	ssvProbeNode           ServerSideInputValidatorFunc
	ssvNodeGetCapabilities ServerSideInputValidatorFunc
)

// NewServerSideInputValidator initializes and returns new
// ServerSideInputValidator instance.
func NewServerSideInputValidator() ServerSideInputValidator {

	return ServerSideInputValidator{
		// Controller
		FMCreateVolume:               ssvCreateVolume,
		FMDeleteVolume:               ssvDeleteVolume,
		FMControllerPublishVolume:    ssvControllerPublishVolume,
		FMControllerUnpublishVolume:  ssvControllerUnpublishVolume,
		FMValidateVolumeCapabilities: ssvValidateVolumeCapabilities,
		FMListVolumes:                ssvListVolumes,
		FMGetCapacity:                ssvGetCapacity,
		FMControllerGetCapabilities:  ssvControllerGetCapabilities,

		// Identity
		FMGetSupportedVersions: ssvGetSupportedVersions,
		FMGetPluginInfo:        ssvGetPluginInfo,

		// Nodess
		FMNodePublishVolume:   ssvNodePublishVolume,
		FMNodeUnpublishVolume: ssvNodeUnpublishVolume,
		FMGetNodeID:           ssvGetNodeID,
		FMProbeNode:           ssvProbeNode,
		FMNodeGetCapabilities: ssvNodeGetCapabilities,
	}
}

// Handle is a unary, server interceptor function.
func (v ServerSideInputValidator) Handle(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	// Lookup the validator for the message's full method. If such
	// a validator function exists then use it to validate the
	// request. If invalid then return the error. Otherwise
	// continue procesing the request.
	if h, ok := v[info.FullMethod]; ok && h != nil {
		resp, err := h(ctx, req)
		if err != nil {
			return nil, err
		}
		if resp != nil {
			return resp, nil
		}
	}
	return handler(ctx, req)
}

type hasGetNameAsString interface {
	GetName() string
}
