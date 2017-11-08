package gocsi

import (
	"fmt"
	"io"
	"regexp"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// NewServerRequestLogger initializes a new unary, server interceptor
// that logs request details.
func NewServerRequestLogger(
	stdout, stderr io.Writer) grpc.UnaryServerInterceptor {

	return (&serverReqLogger{stdout: stdout, stderr: stderr}).handle
}

type serverReqLogger struct {
	stdout io.Writer
	stderr io.Writer
}

var emptyValRX = regexp.MustCompile(
	`^((?:)|(?:\[\])|(?:<nil>)|(?:map\[\]))$`)

func (v *serverReqLogger) handle(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	if req == nil {
		return handler(ctx, req)
	}

	w := v.stdout
	fmt.Fprintf(w, "%s: ", info.FullMethod)
	if rid, ok := GetRequestID(ctx); ok {
		fmt.Fprintf(w, "REQ %04d", rid)
	}
	rprintReqOrRep(w, req)
	fmt.Fprintln(w)

	return handler(ctx, req)
}
