package requestid

import (
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	csictx "github.com/rexray/gocsi/context"
)

// Middleware injects a unique ID into outgoing requests and reads the
// ID from incoming requests.
type Middleware struct {
	sync.Once
	id uint64
}

// Init is available to explicitly initialize the middleware.
func (s *Middleware) Init(ctx context.Context) (err error) {
	return s.initOnce(ctx)
}

func (s *Middleware) initOnce(ctx context.Context) (err error) {
	s.Once.Do(func() {
		err = s.init(ctx)
	})
	return
}

func (s *Middleware) init(ctx context.Context) error {
	log.Info("middleware: request id injection")
	return nil
}

// HandleServer is a UnaryServerInterceptor
// that reads a unique request ID from the incoming context's gRPC
// metadata. If the incoming context does not contain gRPC metadata or
// a request ID, then a new request ID is generated.
func (s *Middleware) HandleServer(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	if err := s.initOnce(ctx); err != nil {
		return nil, err
	}

	// storeID is a flag that indicates whether or not the request ID
	// should be atomically stored in the interceptor's id field at
	// the end of this function. If the ID was found in the incoming
	// request and could be parsed successfully then the ID is stored.
	// If the ID was generated server-side then the ID is not stored.
	storeID := true

	// Retrieve the gRPC metadata from the incoming context.
	md, mdOK := metadata.FromIncomingContext(ctx)

	// If no gRPC metadata was found then create some and ensure the
	// context is a gRPC incoming context.
	if !mdOK {
		md = metadata.Pairs()
		ctx = metadata.NewIncomingContext(ctx, md)
	}

	// Check the metadata from the request ID.
	szID, szIDOK := md[csictx.RequestIDKey]

	// If the metadata does not contain a request ID then create a new
	// request ID and inject it into the metadata.
	if !szIDOK || len(szID) != 1 {
		szID = []string{fmt.Sprintf("%d", atomic.AddUint64(&s.id, 1))}
		md[csictx.RequestIDKey] = szID
		storeID = false
	}

	// Parse the request ID from the
	id, err := strconv.ParseUint(szID[0], 10, 64)
	if err != nil {
		id = atomic.AddUint64(&s.id, 1)
		storeID = false
	}

	if storeID {
		atomic.StoreUint64(&s.id, id)
	}

	return handler(ctx, req)
}

// HandleClient is a UnaryClientInterceptor
// that injects the outgoing context with gRPC metadata that contains
// a unique ID.
func (s *Middleware) HandleClient(
	ctx context.Context,
	method string,
	req, rep interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption) error {

	// Ensure there is an outgoing gRPC context with metadata.
	md, mdOK := metadata.FromOutgoingContext(ctx)
	if !mdOK {
		md = metadata.Pairs()
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	// Ensure the request ID is set in the metadata.
	if szID, szIDOK := md[csictx.RequestIDKey]; !szIDOK || len(szID) != 1 {
		szID = []string{fmt.Sprintf("%d", atomic.AddUint64(&s.id, 1))}
		md[csictx.RequestIDKey] = szID
	}

	return invoker(ctx, method, req, rep, cc, opts...)
}

// Usage returns the middleware's usage string.
func (s *Middleware) Usage() string {
	return `REQUEST ID INJECTION
    X_CSI_REQ_ID_INJECTION
        A flag that enables request ID injection. The ID is parsed from
        the incoming request's metadata with a key of "csi.requestid".
        If no value for that key is found then a new request ID is
        generated using an atomic sequence counter.`
}
