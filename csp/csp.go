package csp

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"text/template"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/thecodeteam/gocsi"
)

// Run launches a CSI storage plug-in.
func Run(
	ctx context.Context,
	appName, appDescription, appUsage string,
	sp StoragePluginProvider) {

	// Check for the debug value.
	if v, ok := gocsi.LookupEnv(ctx, EnvVarDebug); ok {
		if ok, _ := strconv.ParseBool(v); ok {
			gocsi.Setenv(ctx, EnvVarLogLevel, "debug")
			gocsi.Setenv(ctx, EnvVarReqLogging, "true")
			gocsi.Setenv(ctx, EnvVarRepLogging, "true")
		}
	}

	// Adjust the log level.
	var lvl log.Level
	if v, ok := gocsi.LookupEnv(ctx, EnvVarLogLevel); ok {
		var err error
		if lvl, err = log.ParseLevel(v); err != nil {
			lvl = log.WarnLevel
		}
	}
	log.SetLevel(lvl)

	printUsage := func() {
		// app is the information passed to the printUsage function
		app := struct {
			Name        string
			Description string
			Usage       string
			BinPath     string
		}{
			appName,
			appDescription,
			appUsage,
			os.Args[0],
		}

		t, err := template.New("t").Parse(usage)
		if err != nil {
			log.WithError(err).Fatalln("failed to parse usage template")
		}
		if err := t.Execute(os.Stderr, app); err != nil {
			log.WithError(err).Fatalln("failed emitting usage")
		}
		return
	}

	// Check for a help flag.
	fs := flag.NewFlagSet("csp", flag.ExitOnError)
	fs.Usage = printUsage
	var help bool
	fs.BoolVar(&help, "?", false, "")
	err := fs.Parse(os.Args)
	if err == flag.ErrHelp || help {
		printUsage()
		os.Exit(1)
	}

	// If no endpoint is set then print the usage.
	if os.Getenv(EnvVarEndpoint) == "" {
		printUsage()
		os.Exit(1)
	}

	l, err := gocsi.GetCSIEndpointListener()
	if err != nil {
		log.WithError(err).Fatalln("failed to listen")
	}

	// Define a lambda that can be used in the exit handler
	// to remove a potential UNIX sock file.
	rmSockFile := func() {
		if l == nil || l.Addr() == nil {
			return
		}
		if l.Addr().Network() == "unix" {
			sockFile := l.Addr().String()
			os.RemoveAll(sockFile)
			log.WithField("path", sockFile).Info("removed sock file")
		}
	}

	trapSignals(func() {
		sp.GracefulStop(ctx)
		rmSockFile()
		log.Info("server stopped gracefully")
	}, func() {
		sp.Stop(ctx)
		rmSockFile()
		log.Info("server aborted")
	})

	if err := sp.Serve(ctx, l); err != nil {
		log.WithError(err).Fatal("grpc failed")
		os.Exit(1)
	}
}

// StoragePluginProvider is able to serve a gRPC endpoint that provides
// the CSI services: Controller, Identity, Node.
type StoragePluginProvider interface {

	// Serve accepts incoming connections on the listener lis, creating
	// a new ServerTransport and service goroutine for each. The service
	// goroutine read gRPC requests and then call the registered handlers
	// to reply to them. Serve returns when lis.Accept fails with fatal
	// errors.  lis will be closed when this method returns.
	// Serve always returns non-nil error.
	Serve(ctx context.Context, lis net.Listener) error

	// Stop stops the gRPC server. It immediately closes all open
	// connections and listeners.
	// It cancels all active RPCs on the server side and the corresponding
	// pending RPCs on the client side will get notified by connection
	// errors.
	Stop(ctx context.Context)

	// GracefulStop stops the gRPC server gracefully. It stops the server
	// from accepting new connections and RPCs and blocks until all the
	// pending RPCs are finished.
	GracefulStop(ctx context.Context)
}

// StoragePlugin is the collection of services and data used to server
// a new gRPC endpoint that acts as a CSI storage plug-in (SP).
type StoragePlugin struct {
	// Controller is the eponymous CSI service.
	Controller csi.ControllerServer

	// Identity is the eponymous CSI service.
	Identity csi.IdentityServer

	// Node is the eponymous CSI service.
	Node csi.NodeServer

	// ServerOpts is a list of gRPC server options used when serving
	// the SP. This list should not include a gRPC interceptor option
	// as one is created automatically based on the interceptor configuration
	// or provided list of interceptors.
	ServerOpts []grpc.ServerOption

	// Interceptors is a list of gRPC server interceptors to use when
	// serving the SP. This list should not include the interceptors
	// defined in the GoCSI package as those are configured by default
	// based on runtime configuration settings.
	Interceptors []grpc.UnaryServerInterceptor

	// IdempotencyProvider is used to provide a simple way of ensuring the
	// SP is idempotent. If this value nil then the SP is responsible for
	// ensuring idempotency.
	IdempotencyProvider gocsi.IdempotencyProvider

	// BeforeServe is an optional callback that is invoked after the
	// StoragePlugin has been initialized, just prior to the creation
	// of the gRPC server. This callback may be used to perform custom
	// initialization logic, modify the interceptors and server options,
	// or prevent the server from starting by returning a non-nil error.
	BeforeServe func(context.Context, *StoragePlugin, net.Listener) error

	// EnvVars is a list of default environment variables and values.
	EnvVars []string

	serveOnce sync.Once
	stopOnce  sync.Once
	server    *grpc.Server

	envVars           map[string]string
	pluginInfo        csi.GetPluginInfoResponse
	supportedVersions []csi.Version
}

// Serve accepts incoming connections on the listener lis, creating
// a new ServerTransport and service goroutine for each. The service
// goroutine read gRPC requests and then call the registered handlers
// to reply to them. Serve returns when lis.Accept fails with fatal
// errors.  lis will be closed when this method returns.
// Serve always returns non-nil error.
func (sp *StoragePlugin) Serve(ctx context.Context, lis net.Listener) error {
	var err error
	sp.serveOnce.Do(func() {
		// Please note that the order of the below init functions is
		// important and should not be altered unless by someone aware
		// of how they work.

		// Adding this function to the context allows `gocsi.LookupEnv`
		// to search this SP's default env vars for a value.
		ctx = gocsi.WithLookupEnv(ctx, sp.lookupEnv)

		// Initialize the storage plug-in's environment variables map.
		sp.initEnvVars(ctx)

		// Initialize the storage plug-in's list of supported versions.
		sp.initSupportedVersions(ctx)

		// Initialize the storage plug-in's info.
		sp.initPluginInfo(ctx)

		// Initialize the interceptors.
		sp.initInterceptors(ctx)

		// Invoke the SP's BeforeServe function to give the SP a chance
		// to perform any local initialization routines.
		if f := sp.BeforeServe; f != nil {
			if err = f(ctx, sp, lis); err != nil {
				return
			}
		}

		// Add the interceptors to the server if any are configured.
		if i := sp.Interceptors; len(i) > 0 {
			sp.ServerOpts = append(sp.ServerOpts,
				grpc.UnaryInterceptor(gocsi.ChainUnaryServer(i...)))
		}

		// Initialize the gRPC server.
		sp.server = grpc.NewServer(sp.ServerOpts...)

		// Register the CSI services.
		if s := sp.Controller; s != nil {
			csi.RegisterControllerServer(sp.server, s)
		}
		if s := sp.Identity; s != nil {
			csi.RegisterIdentityServer(sp.server, s)
		}
		if s := sp.Node; s != nil {
			csi.RegisterNodeServer(sp.server, s)
		}

		endpoint := fmt.Sprintf(
			"%s://%s",
			lis.Addr().Network(), lis.Addr().String())
		log.WithField("endpoint", endpoint).Info("serving")

		// Start the gRPC server.
		err = sp.server.Serve(lis)
		return
	})
	return err
}

// Stop stops the gRPC server. It immediately closes all open
// connections and listeners.
// It cancels all active RPCs on the server side and the corresponding
// pending RPCs on the client side will get notified by connection
// errors.
func (sp *StoragePlugin) Stop(ctx context.Context) {
	sp.stopOnce.Do(func() {
		sp.server.Stop()
		log.Info("stopped")
	})
}

// GracefulStop stops the gRPC server gracefully. It stops the server
// from accepting new connections and RPCs and blocks until all the
// pending RPCs are finished.
func (sp *StoragePlugin) GracefulStop(ctx context.Context) {
	sp.stopOnce.Do(func() {
		sp.server.GracefulStop()
		log.Info("gracefully stopped")
	})
}

func (sp *StoragePlugin) lookupEnv(key string) (string, bool) {
	val, ok := sp.envVars[key]
	return val, ok
}

func (sp *StoragePlugin) getEnvBool(ctx context.Context, key string) bool {
	v, ok := gocsi.LookupEnv(ctx, key)
	if !ok {
		return false
	}
	if b, err := strconv.ParseBool(v); err == nil {
		return b
	}
	return false
}

func trapSignals(onExit, onAbort func()) {
	sigc := make(chan os.Signal, 1)
	sigs := []os.Signal{
		syscall.SIGTERM,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
	}
	signal.Notify(sigc, sigs...)
	go func() {
		for s := range sigc {
			ok, graceful := isExitSignal(s)
			if !ok {
				continue
			}
			if !graceful {
				log.WithField("signal", s).Error("received signal; aborting")
				if onAbort != nil {
					onAbort()
				}
				os.Exit(1)
			}
			log.WithField("signal", s).Info("received signal; shutting down")
			if onExit != nil {
				onExit()
			}
			os.Exit(0)
		}
	}()
}

// isExitSignal returns a flag indicating whether a signal SIGHUP,
// SIGINT, SIGTERM, or SIGQUIT. The second return value is whether it is a
// graceful exit. This flag is true for SIGTERM, SIGHUP, SIGINT, and SIGQUIT.
func isExitSignal(s os.Signal) (bool, bool) {
	switch s {
	case syscall.SIGTERM,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT:
		return true, true
	default:
		return false, false
	}
}

type logger struct {
	f func(msg string, args ...interface{})
	w io.Writer
}

func newLogger(f func(msg string, args ...interface{})) *logger {
	l := &logger{f: f}
	r, w := io.Pipe()
	l.w = w
	go func() {
		scan := bufio.NewScanner(r)
		for scan.Scan() {
			f(scan.Text())
		}
	}()
	return l
}

func (l *logger) Write(data []byte) (int, error) {
	return l.w.Write(data)
}
