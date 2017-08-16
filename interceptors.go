package gocsi

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"sync/atomic"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
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
