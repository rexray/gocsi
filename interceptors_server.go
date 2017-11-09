package gocsi

import (
	"fmt"
	"io"
	"reflect"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

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

type listVolsPager struct {
}

type nextVolToken struct {
	ID    string `json:"id"`
	Token uint32 `json:"tok"`
}

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
		if name == "UserCredentials" {
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
