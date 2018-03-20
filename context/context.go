package context

import (
	"context"
	"strconv"

	"google.golang.org/grpc/metadata"
)

// RequestIDKey is the key used to put/get a CSI request ID
// in/fromt a Go context.
const RequestIDKey = "csi.requestid"

var (
	// ctxRequestIDKey is an interface-wrapped key used to access the
	// gRPC request ID injected into an outgoing or incoming context
	// via the GoCSI request ID injection interceptor
	ctxRequestIDKey = interface{}("x-csi-request-id")
)

// GetRequestID inspects the context for gRPC metadata and returns
// its request ID if available.
func GetRequestID(ctx context.Context) (uint64, bool) {
	var (
		szID   []string
		szIDOK bool
	)

	// Prefer the incoming context, but look in both types.
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		szID, szIDOK = md[RequestIDKey]
	} else if md, ok := metadata.FromOutgoingContext(ctx); ok {
		szID, szIDOK = md[RequestIDKey]
	}

	if szIDOK && len(szID) == 1 {
		if id, err := strconv.ParseUint(szID[0], 10, 64); err == nil {
			return id, true
		}
	}

	return 0, false
}
