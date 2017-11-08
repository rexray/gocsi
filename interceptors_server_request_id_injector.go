package gocsi

import (
	"sync/atomic"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var requestIDVal uint64

// ServerRequestIDInjector is a unary server interceptor that injects
// request contexts with a unique request ID.
func ServerRequestIDInjector(
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
