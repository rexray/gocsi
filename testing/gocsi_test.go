package gocsi_test

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/akutz/memconn"
	"github.com/onsi/ginkgo"
	gomegaTypes "github.com/onsi/gomega/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dell/gocsi/mock/provider"
)

func startMockServer(ctx context.Context) (*grpc.ClientConn, func(), error) {

	// Create a new Mock SP instance and serve it with a piped connection.
	sp := provider.New()
	lis, err := memconn.Listen("memu", "csi-test")
	Ω(err).Should(BeNil())
	go func() {
		defer GinkgoRecover()
		if err := sp.Serve(ctx, lis); err != nil {
			Ω(err.Error()).Should(Equal("http: Server closed"))
		}
	}()

	clientOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDialer(func(string, time.Duration) (net.Conn, error) {
			return memconn.Dial("memu", "csi-test")
		}),
	}

	// Create a client for the piped connection.
	client, err := grpc.DialContext(ctx, "", clientOpts...)
	Ω(err).ShouldNot(HaveOccurred())

	return client, func() { sp.GracefulStop(ctx) }, nil
}

// CTest is an alias to retrieve the current Ginko test description.
var CTest = ginkgo.CurrentGinkgoTestDescription

type grpcErrorMatcher struct {
	exp error
}

func (m *grpcErrorMatcher) Match(actual interface{}) (bool, error) {
	statExp, ok := status.FromError(m.exp)
	if !ok {
		return false, fmt.Errorf(
			"expected error not gRPC error: %T", m.exp)
	}

	actErr, ok := actual.(error)
	if !ok {
		return false, fmt.Errorf(
			"invalid actual error: %T", actual)
	}

	statAct, ok := status.FromError(actErr)
	if !ok {
		return false, fmt.Errorf(
			"actual error not gRPC error: %T", actual)
	}

	if statExp.Code() != statAct.Code() {
		return false, nil
	}

	if statExp.Message() != statAct.Message() {
		return false, nil
	}

	return true, nil
}
func (m *grpcErrorMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf(
		"Expected\n\t%#v\nto be equal to\n\t%#v", actual, m.exp)
}
func (m *grpcErrorMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf(
		"Expected\n\t%#v\nnot to be equal to\n\t%#v", actual, m.exp)
}

// Σ is a custom Ginkgo matcher that compares two gRPC errors.
func Σ(a error) gomegaTypes.GomegaMatcher {
	return &grpcErrorMatcher{exp: a}
}

// ΣCM is a custom Ginkgo matcher that compares two gRPC errors.
func ΣCM(c codes.Code, m string, args ...interface{}) gomegaTypes.GomegaMatcher {
	return &grpcErrorMatcher{exp: status.Errorf(c, m, args...)}
}

const string128 = "0000000000000000000000000000000000000000000000000000000000" +
	"0000000000000000000000000000000000000000000000000000000000000000000000"
const string129 = "0000000000000000000000000000000000000000000000000000000000" +
	"00000000000000000000000000000000000000000000000000000000000000000000000"
