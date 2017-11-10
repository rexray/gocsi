package gocsi

import (
	"sync/atomic"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type serverRequestIDInjector struct {
	requestIDVal uint64
}

// NewServerRequestIDInjector returns a new server interceptor that injects
// request contexts with a unique request ID.
func NewServerRequestIDInjector() grpc.UnaryServerInterceptor {
	return (&serverRequestIDInjector{}).handle
}

func (i *serverRequestIDInjector) handle(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	return handler(
		context.WithValue(
			ctx,
			requestIDKey,
			atomic.AddUint64(&i.requestIDVal, 1)),
		req)
}
