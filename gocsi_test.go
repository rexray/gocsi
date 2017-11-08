package gocsi_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/onsi/ginkgo"
	gomegaTypes "github.com/onsi/gomega/types"
	"google.golang.org/grpc"

	"github.com/thecodeteam/gocsi"
	"github.com/thecodeteam/gocsi/csi"
	"github.com/thecodeteam/gocsi/mock/provider"
)

func startMockServer(ctx context.Context) (*grpc.ClientConn, func(), error) {

	// Create a new Mock SP instance and serve it with a piped connection.
	sp := provider.New()
	pipeconn := gocsi.NewPipeConn("csi-test")
	go func() {
		if err := sp.Serve(ctx, pipeconn); err != nil {
			Ω(err.Error()).Should(Equal("http: Server closed"))
		}
	}()

	// Create a client for the piped connection.
	client, err := grpc.DialContext(
		ctx, "",
		grpc.WithInsecure(),
		grpc.WithDialer(pipeconn.DialGrpc),
		grpc.WithUnaryInterceptor(gocsi.ChainUnaryClient(
			gocsi.ClientCheckReponseError,
			gocsi.NewClientResponseValidator())))
	Ω(err).ShouldNot(HaveOccurred())

	return client, func() { sp.GracefulStop(ctx) }, nil
}

func newCSIVersion(major, minor, patch uint32) *csi.Version {
	return &csi.Version{
		Major: major,
		Minor: minor,
		Patch: patch,
	}
}

var mockSupportedVersions = []*csi.Version{
	newCSIVersion(0, 1, 0),
	newCSIVersion(0, 2, 0),
	newCSIVersion(1, 0, 0),
	newCSIVersion(1, 1, 0),
}

// CTest is an alias to retrieve the current Ginko test description.
var CTest = ginkgo.CurrentGinkgoTestDescription

type gocsiErrMatcher struct {
	exp *gocsi.Error
}

func (m *gocsiErrMatcher) Match(actual interface{}) (bool, error) {
	act, ok := actual.(*gocsi.Error)
	if !ok {
		return false, errors.New("gocsiErrMatcher expects a *gocsi.Error")
	}
	if m.exp.Code != act.Code {
		return false, nil
	}
	if m.exp.Description != act.Description {
		return false, nil
	}
	if m.exp.FullMethod != act.FullMethod {
		return false, nil
	}
	return true, nil
}
func (m *gocsiErrMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf(
		"Expected\n\t%#v\nto be equal to\n\t%#v", actual, m.exp)
}
func (m *gocsiErrMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf(
		"Expected\n\t%#v\nnot to be equal to\n\t%#v", actual, m.exp)
}

// Σ is a custom Ginkgo matcher that compares two GoCSI errors.
func Σ(a *gocsi.Error) gomegaTypes.GomegaMatcher {
	return &gocsiErrMatcher{exp: a}
}
