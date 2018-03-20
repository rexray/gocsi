package logging

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	csictx "github.com/rexray/gocsi/context"
	csienv "github.com/rexray/gocsi/env"
	"github.com/rexray/gocsi/utils"
)

// Middleware provides logging capabilities for gRPC requests and responses.
type Middleware struct {
	sync.Once

	// RequestWriter is the request writer.
	RequestWriter io.Writer

	// ResponseWriter is the response writer.
	ResponseWriter io.Writer
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
	if csienv.IsDebug(ctx) {
		if s.RequestWriter == nil {
			s.RequestWriter = os.Stdout
		}
		if s.ResponseWriter == nil {
			s.ResponseWriter = os.Stdout
		}
	}

	if v, ok := csienv.LookupEnv(ctx, "X_CSI_REQ_LOGGING"); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			if b && s.RequestWriter == nil {
				s.RequestWriter = os.Stdout
			} else {
				s.RequestWriter = nil
			}
		}
	}
	if v, ok := csienv.LookupEnv(ctx, "X_CSI_REP_LOGGING"); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			if b && s.ResponseWriter == nil {
				s.ResponseWriter = os.Stdout
			} else {
				s.ResponseWriter = nil
			}
		}
	}

	if s.RequestWriter != nil {
		log.Info("middleware: request logging")
	}
	if s.ResponseWriter != nil {
		log.Info("middleware: response logging")
	}

	return nil
}

// HandleServer is a server-side, unary gRPC interceptor.
func (s *Middleware) HandleServer(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	return s.handle(ctx, info.FullMethod, req, func() (interface{}, error) {
		return handler(ctx, req)
	})
}

// HandleClient is a client-side, unary gRPC interceptor.
func (s *Middleware) HandleClient(
	ctx context.Context,
	method string,
	req, rep interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption) error {

	_, err := s.handle(ctx, method, req, func() (interface{}, error) {
		return rep, invoker(ctx, method, req, rep, cc, opts...)
	})
	return err
}

func (s *Middleware) handle(
	ctx context.Context,
	method string,
	req interface{},
	next func() (interface{}, error)) (rep interface{}, failed error) {

	if err := s.initOnce(ctx); err != nil {
		return nil, err
	}

	// If the request is nil then pass control to the next handler
	// in the chain.
	if req == nil {
		return next()
	}

	w := &bytes.Buffer{}
	reqID, reqIDOK := csictx.GetRequestID(ctx)

	// Print the request
	if s.RequestWriter != nil {
		fmt.Fprintf(w, "%s: ", method)
		if reqIDOK {
			fmt.Fprintf(w, "REQ %04d", reqID)
		}
		rprintReqOrRep(w, req)
		fmt.Fprintln(s.RequestWriter, w.String())
	}

	w.Reset()

	// Get the response.
	rep, failed = next()

	if s.ResponseWriter == nil {
		return
	}

	// Print the response method name.
	fmt.Fprintf(w, "%s: ", method)
	if reqIDOK {
		fmt.Fprintf(w, "REP %04d", reqID)
	}

	// Print the response error if it is set.
	if failed != nil {
		fmt.Fprint(w, ": ")
		fmt.Fprint(w, failed)
	}

	// Print the response data if it is set.
	if !utils.IsNilResponse(rep) {
		rprintReqOrRep(w, rep)
	}
	fmt.Fprintln(s.ResponseWriter, w.String())

	return
}

var emptyValRX = regexp.MustCompile(
	`^((?:)|(?:\[\])|(?:<nil>)|(?:map\[\]))$`)

// rprintReqOrRep is used by the server-side interceptors that log
// requests and responses.
func rprintReqOrRep(w io.Writer, obj interface{}) {
	rv := reflect.ValueOf(obj).Elem()
	tv := rv.Type()
	nf := tv.NumField()
	printedColon := false
	printComma := false
	for i := 0; i < nf; i++ {
		name := tv.Field(i).Name
		if strings.Contains(name, "Secrets") {
			continue
		}
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
		fmt.Fprintf(w, "%s=%s", name, sv)
	}
}

// Usage returns the middleware's usage string.
func (s *Middleware) Usage() string {
	return usage
}

const usage = `REQUEST & RESPONSE LOGGING
    X_CSI_REQ_LOGGING
        A flag that enables logging of incoming requests to STDOUT.

    X_CSI_REP_LOGGING
        A flag that enables logging of outgoing responses to STDOUT.`

func getEnvBool(ctx context.Context, key string) bool {
	v, ok := csienv.LookupEnv(ctx, key)
	if !ok {
		return false
	}
	if b, err := strconv.ParseBool(v); err == nil {
		return b
	}
	return false
}
