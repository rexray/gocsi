//go:generate make

// Package gocsi provides a Container Storage Interface (CSI) library,
// client, and other helpful utilities.
package gocsi

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"os/user"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"text/template"

	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	csienv "github.com/rexray/gocsi/env"
	"github.com/rexray/gocsi/utils"
)

// Run launches a CSI storage plug-in.
func Run(
	ctx context.Context,
	appName, appDescription, appUsage string,
	sp StoragePluginProvider) {

	// Adjust the log level.
	lvl := log.InfoLevel
	if csienv.IsDebug(ctx) {
		lvl = log.DebugLevel
	} else {
		if v, ok := csienv.LookupEnv(ctx, csienv.LogLevel); ok {
			if lvl2, err := log.ParseLevel(v); err == nil {
				lvl = lvl2
			}
		}
	}
	log.SetLevel(lvl)

	var spUsage string
	if o, ok := sp.(hasUsage); ok {
		spUsage = o.Usage()
	}

	printUsage := func() {
		// app is the information passed to the printUsage function
		app := struct {
			Name               string
			Description        string
			AppUsage           string
			StoragePluginUsage string
			BinPath            string
		}{
			appName,
			appDescription,
			appUsage,
			spUsage,
			os.Args[0],
		}

		t, err := template.New("t").Parse(usageTemplate)
		if err != nil {
			log.WithError(err).Fatalln("failed to parse usage template")
		}
		if err := t.Execute(os.Stderr, app); err != nil {
			log.WithError(err).Fatalln("failed emitting usage")
		}
		return
	}

	// Check for a help flag.
	var help bool
	flag.Usage = printUsage
	flag.BoolVar(&help, "?", false, "")
	flag.Parse()
	if help {
		printUsage()
		os.Exit(1)
	}

	// If no endpoint is set then print the usage.
	if os.Getenv(csienv.Endpoint) == "" {
		printUsage()
		os.Exit(1)
	}

	l, err := utils.GetCSIEndpointListener()
	if err != nil {
		log.WithError(err).Fatalln("failed to listen")
	}

	// Define a lambda that can be used in the exit handler
	// to remove a potential UNIX sock file.
	var rmSockFileOnce sync.Once
	rmSockFile := func() {
		rmSockFileOnce.Do(func() {
			if l == nil || l.Addr() == nil {
				return
			}
			if l.Addr().Network() == netUnix {
				sockFile := l.Addr().String()
				os.RemoveAll(sockFile)
				log.WithField("path", sockFile).Info("removed sock file")
			}
		})
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
		rmSockFile()
		log.WithError(err).Fatal("grpc failed")
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

	// Middleware is a list of gRPC server-side middleware to use when
	// serving the SP.
	Middleware []ServerMiddleware

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

	envVars map[string]string
}

// Serve accepts incoming connections on the listener lis, creating
// a new ServerTransport and service goroutine for each. The service
// goroutine read gRPC requests and then call the registered handlers
// to reply to them. Serve returns when lis.Accept fails with fatal
// errors.  lis will be closed when this method returns.
// Serve always returns non-nil error.
func (sp *StoragePlugin) Serve(
	ctx context.Context, lis net.Listener) (err error) {

	sp.serveOnce.Do(func() {
		// Please note that the order of the below init functions is
		// important and should not be altered unless by someone aware
		// of how they work.

		// Adding this function to the context allows `csienv.LookupEnv`
		// to search this SP's default env vars for a value.
		ctx = csienv.WithLookupEnv(ctx, sp.lookupEnv)

		// Adding this function to the context allows `csienv.Setenv`
		// to set environment variables in this SP's env var store.
		ctx = csienv.WithSetenv(ctx, sp.setEnv)

		// Initialize the storage plug-in's environment variables map.
		sp.initEnvVars(ctx)

		// Adjust the endpoint's file permissions.
		if err = sp.initEndpointPerms(ctx, lis); err != nil {
			return
		}

		// Adjust the endpoint's file ownership.
		if err = sp.initEndpointOwner(ctx, lis); err != nil {
			return
		}

		// Initialize the interceptors.
		if err = sp.initMiddleware(ctx); err != nil {
			return
		}

		// Invoke the SP's BeforeServe function to give the SP a chance
		// to perform any local initialization routines.
		if sp.BeforeServe != nil {
			if err = sp.BeforeServe(ctx, sp, lis); err != nil {
				return
			}
		}

		// Initialize the gRPC server.
		sp.server = grpc.NewServer(sp.ServerOpts...)

		// Register the CSI services.
		// Always require the identity service.
		if sp.Identity == nil {
			err = errors.New("identity service is required")
			return
		}
		// Either a Controller or Node service should be supplied.
		if sp.Controller == nil && sp.Node == nil {
			err = errors.New(
				"either a controller or node service is required")
			return
		}

		// Always register the identity service.
		csi.RegisterIdentityServer(sp.server, sp.Identity)
		log.Info("identity service registered")

		// Determine which of the controller/node services to register
		mode := csienv.Getenv(ctx, csienv.Mode)
		if strings.EqualFold(mode, "controller") {
			mode = "controller"
		} else if strings.EqualFold(mode, "node") {
			mode = "node"
		} else {
			mode = ""
		}

		if mode == "" || mode == "controller" {
			if sp.Controller == nil {
				err = errors.New("controller service is required")
				return
			}
			csi.RegisterControllerServer(sp.server, sp.Controller)
			log.Info("controller service registered")
		}
		if mode == "" || mode == "node" {
			if sp.Node == nil {
				err = errors.New("node service is required")
				return
			}
			csi.RegisterNodeServer(sp.server, sp.Node)
			log.Info("node service registered")
		}

		endpoint := fmt.Sprintf(
			"%s://%s",
			lis.Addr().Network(), lis.Addr().String())
		log.WithField("endpoint", endpoint).Info("serving")

		// Start the gRPC server.
		err = sp.server.Serve(lis)
		return
	})
	return
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

const netUnix = "unix"

func (sp *StoragePlugin) initEndpointPerms(
	ctx context.Context, lis net.Listener) error {

	if lis.Addr().Network() != netUnix {
		return nil
	}

	v, ok := csienv.LookupEnv(ctx, csienv.EndpointPerms)
	if !ok || v == "0755" {
		return nil
	}
	u, err := strconv.ParseUint(v, 8, 32)
	if err != nil {
		return err
	}

	p := lis.Addr().String()
	m := os.FileMode(u)

	log.WithFields(map[string]interface{}{
		"path": p,
		"mode": m,
	}).Info("chmod csi endpoint")

	if err := os.Chmod(p, m); err != nil {
		return err
	}

	return nil
}

func (sp *StoragePlugin) initEndpointOwner(
	ctx context.Context, lis net.Listener) error {

	if lis.Addr().Network() != netUnix {
		return nil
	}

	var (
		usrName string
		grpName string

		uid  = os.Getuid()
		gid  = os.Getgid()
		puid = uid
		pgid = gid
	)

	if v, ok := csienv.LookupEnv(ctx, csienv.EndpointUser); ok {
		m, err := regexp.MatchString(`^\d+$`, v)
		if err != nil {
			return err
		}
		usrName = v
		szUID := v
		if m {
			u, err := user.LookupId(v)
			if err != nil {
				return err
			}
			usrName = u.Username
		} else {
			u, err := user.Lookup(v)
			if err != nil {
				return err
			}
			szUID = u.Uid
		}
		iuid, err := strconv.Atoi(szUID)
		if err != nil {
			return err
		}
		uid = iuid
	}

	if v, ok := csienv.LookupEnv(ctx, csienv.EndpointGroup); ok {
		m, err := regexp.MatchString(`^\d+$`, v)
		if err != nil {
			return err
		}
		grpName = v
		szGID := v
		if m {
			u, err := user.LookupGroupId(v)
			if err != nil {
				return err
			}
			grpName = u.Name
		} else {
			u, err := user.LookupGroup(v)
			if err != nil {
				return err
			}
			szGID = u.Gid
		}
		igid, err := strconv.Atoi(szGID)
		if err != nil {
			return err
		}
		gid = igid
	}

	if uid != puid || gid != pgid {
		f := lis.Addr().String()
		log.WithFields(map[string]interface{}{
			"uid":  usrName,
			"gid":  grpName,
			"path": f,
		}).Info("chown csi endpoint")
		if err := os.Chown(f, uid, gid); err != nil {
			return err
		}
	}

	return nil
}

func (sp *StoragePlugin) initEnvVars(ctx context.Context) {

	// Copy the environment variables from the public EnvVar
	// string slice to the private envVars map for quick lookup.
	sp.envVars = map[string]string{}
	for _, v := range sp.EnvVars {
		// Environment variables must adhere to one of the following
		// formats:
		//
		//     - ENV_VAR_KEY=
		//     - ENV_VAR_KEY=ENV_VAR_VAL
		pair := strings.SplitN(v, "=", 2)
		if len(pair) < 1 || len(pair) > 2 {
			continue
		}

		// Ensure the environment variable is stored in all upper-case
		// to make subsequent map-lookups deterministic.
		key := strings.ToUpper(pair[0])

		// Check to see if the value for the key is available from the
		// context's os.Environ or os.LookupEnv functions. If neither
		// return a value then use the provided default value.
		var val string
		if v, ok := csienv.LookupEnv(ctx, key); ok {
			val = v
		} else if len(pair) > 1 {
			val = pair[1]
		}
		sp.envVars[key] = val
	}

	// If there is an environment variable string slice in the context, be sure
	// to add it to the list of the SP's environment variables.
	if envVars, ok := csienv.GetEnviron(ctx); ok {
		for _, v := range envVars {
			// Environment variables must adhere to one of the following
			// formats:
			//
			//     - ENV_VAR_KEY=
			//     - ENV_VAR_KEY=ENV_VAR_VAL
			pair := strings.SplitN(v, "=", 2)
			if len(pair) < 1 || len(pair) > 2 {
				continue
			}

			// Ensure the environment variable is stored in all upper-case
			// to make subsequent map-lookups deterministic.
			var val string
			key := strings.ToUpper(pair[0])
			if len(pair) > 1 {
				val = pair[1]
			}
			sp.envVars[key] = val
		}
	}

	return
}

func (sp *StoragePlugin) lookupEnv(key string) (string, bool) {
	val, ok := sp.envVars[key]
	return val, ok
}

func (sp *StoragePlugin) setEnv(key, val string) error {
	sp.envVars[key] = val
	return nil
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
