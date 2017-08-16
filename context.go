package gocsi

import (
	"golang.org/x/net/context"
)

type requestIDKeyType uint64

var requestIDKey interface{} = requestIDKeyType(0)

// GetRequestID gets the gRPC request ID from the provided context.
func GetRequestID(ctx context.Context) (uint64, bool) {
	v, ok := ctx.Value(requestIDKey).(uint64)
	return v, ok
}
