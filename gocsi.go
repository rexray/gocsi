//go:generate make

// Package gocsi provides a Container Storage Interface (CSI) library,
// client, and other helpful utilities.
package gocsi

import (
	"net"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	// Namespace is the namesapce used by the protobuf.
	Namespace = "csi"

	// CSIEndpoint is the name of the environment variable that
	// contains the CSI endpoint.
	CSIEndpoint = "CSI_ENDPOINT"
)

// NewGrpcClient returns a new gRPC client connection.
//
// For information on how the endpoint argument please see
// NewGrpcClientWithOpts.
func NewGrpcClient(
	ctx context.Context,
	endpoint string,
	insecure bool) (*grpc.ClientConn, error) {

	dialOpts := []grpc.DialOption{}
	if insecure {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}
	return NewGrpcClientWithOpts(ctx, endpoint, dialOpts...)
}

// NewGrpcClientWithOpts returns a new gRPC client connection using
// the provided dial options.
//
// Do not provide a WithDialer option as one is created using the
// specified endpoint. The endpoint should be formatted as a Go network
// address. For example:
//
//   - tcp://127.0.0.1:7979
//   - unix:///tmp/csi.sock
func NewGrpcClientWithOpts(
	ctx context.Context,
	endpoint string,
	dialOpts ...grpc.DialOption) (*grpc.ClientConn, error) {

	if dialOpts == nil {
		dialOpts = []grpc.DialOption{}
	}

	dialOpts = append(dialOpts,
		grpc.WithDialer(
			func(target string, timeout time.Duration) (net.Conn, error) {
				proto, addr, err := ParseProtoAddr(target)
				if err != nil {
					return nil, err
				}
				return net.DialTimeout(proto, addr, timeout)
			}))

	return grpc.DialContext(ctx, endpoint, dialOpts...)
}
