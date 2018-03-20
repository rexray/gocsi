package gocsi

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	csienv "github.com/rexray/gocsi/env"
	"github.com/rexray/gocsi/utils"
)

// ClientMiddleware is a Middleware type that impelements a client-side
// gRPC interceptor.
type ClientMiddleware interface {
	HandleClient(
		ctx context.Context,
		method string,
		req, rep interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption) error
}

// ServerMiddleware is a Middleware type that implements a server-side
// gRPC intercetor.
type ServerMiddleware interface {
	HandleServer(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error)
}

// ClientServerMiddleware is a Middleware type that implements both
// client-side and server-side interceptors.
type ClientServerMiddleware interface {
	ClientMiddleware
	ServerMiddleware
}

type hasInit interface {
	Init(context.Context) error
}

func (sp *StoragePlugin) initMiddleware(ctx context.Context) error {
	interceptors := make([]grpc.UnaryServerInterceptor, len(sp.Middleware)+1)
	interceptors[0] = sp.injectContext

	// Range over and initialize the provided middleware.
	for i, j := range sp.Middleware {
		if o, ok := j.(hasInit); ok {
			if err := o.Init(ctx); err != nil {
				return err
			}
		}
		interceptors[i+1] = j.HandleServer
	}

	sp.ServerOpts = append(sp.ServerOpts,
		grpc.UnaryInterceptor(utils.ChainUnaryServer(interceptors...)))

	return nil
}

func (sp *StoragePlugin) injectContext(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	return handler(csienv.WithLookupEnv(ctx, sp.lookupEnv), req)
}
