package plugininfo

import (
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"

	csienv "github.com/rexray/gocsi/env"
	"github.com/rexray/gocsi/utils"
)

// Middleware is server-side middleware that handles the
// GetPluginInfo RPC.
type Middleware struct {
	sync.Once
	Name          string
	VendorVersion string
	Manifest      map[string]string
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
	szInfo, ok := csienv.LookupEnv(ctx, "X_CSI_PLUGIN_INFO")
	if !ok {
		return nil
	}
	info := strings.SplitN(szInfo, ",", 3)

	if len(info) == 0 {
		return nil
	}
	fields := map[string]interface{}{}
	s.Name = strings.TrimSpace(info[0])
	fields["name"] = s.Name

	if len(info) > 1 {
		s.VendorVersion = strings.TrimSpace(info[1])
		fields["vendorVersion"] = s.VendorVersion
	}
	if len(info) > 2 {
		s.Manifest = utils.ParseMap(strings.TrimSpace(info[2]))
		fields["manifest"] = s.Manifest
	}

	log.WithFields(fields).Info("middleware: plug-in info")
	return nil
}

// HandleServer is a server-side, unary gRPC interceptor.
func (s *Middleware) HandleServer(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	if err := s.initOnce(ctx); err != nil {
		return nil, err
	}

	if s.Name == "" {
		return handler(ctx, req)
	}

	_, service, method, err := utils.ParseMethod(info.FullMethod)
	if err != nil {
		return nil, err
	}
	if service == "Identity" && method == "GetPluginInfo" {
		return &csi.GetPluginInfoResponse{
			Name:          s.Name,
			VendorVersion: s.VendorVersion,
			Manifest:      s.Manifest,
		}, nil
	}
	return handler(ctx, req)
}

// Usage returns the middleware's usage string.
func (s *Middleware) Usage() string {
	return usage
}

const usage = `PLUG-IN INFO HANDLER
    X_CSI_PLUGIN_INFO
        The plug-in information is specified via the following
        comma-separated format:

            NAME, VENDOR_VERSION[, MANIFEST...]

        The MANIFEST value may be a series of additional
        comma-separated key/value pairs.

        Please see the encoding/csv package (https://goo.gl/1j1xb9) for
        information on how to quote keys and/or values to include
        leading and trailing whitespace.

        Setting this environment variable will cause the program to
        bypass the SP's GetPluginInfo RPC and returns the specified
        information instead.`
