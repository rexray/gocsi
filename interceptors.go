package gocsi

import (
	"strings"

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

type hasGetName interface {
	GetName() string
}

// ValidateCreateVolume is a gRPC interceptor that validates the
// arguments for the eponymous message.
func ValidateCreateVolume(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp interface{}, err error) {

	if !strings.HasSuffix(info.FullMethod, "/CreateVolume") {
		return handler(ctx, req)
	}

	if req.(hasGetName).GetName() == "" {
		// INVALID_VOLUME_NAME
		return ErrCreateVolume(3, "missing name"), nil
	}

	return handler(ctx, req)
}
