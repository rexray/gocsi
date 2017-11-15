package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/thecodeteam/gocsi"
)

var root struct {
	ctx       context.Context
	client    *grpc.ClientConn
	tpl       *template.Template
	userCreds map[string]string

	logLevel string
	format   string
	endpoint string
	insecure bool
	timeout  time.Duration
	version  csiVersionArg
	metadata mapOfStringArg

	withReqLogging bool
	withRepLogging bool

	withSpecValidator                    bool
	withRequiresCreds                    bool
	withSuccessCreateVolumeAlreadyExists bool
	withSuccessDeleteVolumeNotFound      bool
	withRequiresNodeID                   bool
	withRequiresPubVolInfo               bool
	withRequiresVolumeAttributes         bool
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "csc",
	Short: "a command line client for csi storage plug-ins",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

		ll, err := log.ParseLevel(root.logLevel)
		if err != nil {
			return fmt.Errorf("invalid log level: %v: %v", root.logLevel, err)
		}
		log.SetLevel(ll)

		root.ctx = context.Background()
		log.Debug("assigned the root context")

		// Initialize the template if necessary.
		if root.format == "" {
			switch cmd.Name() {
			case listVolumesCmd.Name():
				if listVolumes.paging {
					root.format = volumeInfoFormat
				} else {
					root.format = listVolumesFormat
				}
			case createVolumeCmd.Name():
				root.format = volumeInfoFormat
			case supportedVersCmd.Name():
				root.format = supportedVersionsFormat
			case pluginInfoCmd.Name():
				root.format = pluginInfoFormat
			}
		}
		if root.format != "" {
			tpl, err := template.New("t").Parse(root.format)
			if err != nil {
				return err
			}
			root.tpl = tpl
		}

		// Parse the credentials if they exist.
		root.userCreds = gocsi.ParseMap(os.Getenv(userCredsKey))

		// Create the gRPC client connection.
		opts := []grpc.DialOption{
			grpc.WithDialer(
				func(target string, timeout time.Duration) (net.Conn, error) {
					proto, addr, err := gocsi.ParseProtoAddr(target)
					if err != nil {
						return nil, err
					}
					return net.DialTimeout(proto, addr, timeout)
				}),
		}

		// Disable TLS if specified.
		if root.insecure {
			opts = append(opts, grpc.WithInsecure())
		}

		var iceptors []grpc.UnaryClientInterceptor

		// Configure logging.
		if root.withReqLogging || root.withRepLogging {

			// Automatically enable request ID injection if logging
			// is enabled.
			iceptors = append(iceptors,
				gocsi.NewClientRequestIDInjector())
			log.Debug("enable request ID injector")

			var (
				loggingOpts []gocsi.LoggingOption
				lout        = newLogger(log.Infof)
			)
			if root.withReqLogging {
				loggingOpts = append(loggingOpts,
					gocsi.WithRequestLogging(lout))
				log.Debug("enable request logging")
			}
			if root.withRepLogging {
				loggingOpts = append(loggingOpts,
					gocsi.WithResponseLogging(lout))
				log.Debug("enable response logging")
			}
			iceptors = append(iceptors,
				gocsi.NewClientLogger(loggingOpts...))
		}

		// Configure the spec validator.
		root.withSpecValidator = root.withSpecValidator ||
			root.withRequiresCreds ||
			root.withSuccessCreateVolumeAlreadyExists ||
			root.withSuccessDeleteVolumeNotFound ||
			root.withRequiresNodeID ||
			root.withRequiresPubVolInfo ||
			root.withRequiresVolumeAttributes
		if root.withSpecValidator {
			var specOpts []gocsi.SpecValidatorOption
			if root.withRequiresCreds {
				specOpts = append(specOpts,
					gocsi.WithRequiresCreateVolumeCredentials(),
					gocsi.WithRequiresDeleteVolumeCredentials(),
					gocsi.WithRequiresControllerPublishVolumeCredentials(),
					gocsi.WithRequiresControllerUnpublishVolumeCredentials(),
					gocsi.WithRequiresNodePublishVolumeCredentials(),
					gocsi.WithRequiresNodeUnpublishVolumeCredentials())
				log.Debug("enable spec validator opt: requires creds")
			}
			if root.withRequiresNodeID {
				specOpts = append(specOpts,
					gocsi.WithRequiresNodeID())
				log.Debug("enable spec validator opt: requires node ID")
			}
			if root.withRequiresPubVolInfo {
				specOpts = append(specOpts,
					gocsi.WithRequiresPublishVolumeInfo())
				log.Debug("enable spec validator opt: requires pub vol info")
			}
			if root.withRequiresVolumeAttributes {
				specOpts = append(specOpts,
					gocsi.WithRequiresVolumeAttributes())
				log.Debug("enable spec validator opt: requires vol attribs")
			}
			if root.withSuccessCreateVolumeAlreadyExists {
				specOpts = append(specOpts,
					gocsi.WithSuccessCreateVolumeAlreadyExists())
				log.Debug("enable spec validator opt: create exists success")
			}
			if root.withSuccessDeleteVolumeNotFound {
				specOpts = append(specOpts,
					gocsi.WithSuccessDeleteVolumeNotFound())
				log.Debug("enable spec validator opt: delete !exists success")
			}
			iceptors = append(iceptors,
				gocsi.NewClientSpecValidator(specOpts...))
		}

		// Add interceptors to the client if any are configured.
		if len(iceptors) > 0 {
			opts = append(opts,
				grpc.WithUnaryInterceptor(gocsi.ChainUnaryClient(iceptors...)))
		}

		ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
		defer cancel()
		client, err := grpc.DialContext(ctx, root.endpoint, opts...)
		if err != nil {
			return err
		}
		root.client = client

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVarP(
		&root.logLevel,
		"log-level",
		"l",
		"warn",
		"the log level")

	RootCmd.PersistentFlags().StringVarP(
		&root.endpoint,
		"endpoint",
		"e",
		os.Getenv("CSI_ENDPOINT"),
		"the csi endpoint")

	RootCmd.PersistentFlags().DurationVarP(
		&root.timeout,
		"timeout",
		"t",
		time.Duration(60)*time.Second,
		"the timeout used for dialing the csi endpoint and invoking rpcs")

	RootCmd.PersistentFlags().BoolVarP(
		&root.insecure,
		"insecure",
		"i",
		true,
		"a flag that disables tls")

	RootCmd.PersistentFlags().VarP(
		&root.metadata,
		"metadata",
		"m",
		"one or more key/value pairs used as grpc metadata")

	RootCmd.PersistentFlags().VarP(
		&root.version,
		"version",
		"v",
		"the csi version to send with an rpc")

	RootCmd.PersistentFlags().BoolVar(
		&root.withReqLogging,
		"with-request-logging",
		false,
		"enables request logging")

	RootCmd.PersistentFlags().BoolVar(
		&root.withRepLogging,
		"with-response-logging",
		false,
		"enables response logging")

	RootCmd.PersistentFlags().BoolVar(
		&root.withSpecValidator,
		"with-spec-validation",
		false,
		"enables validation of request/response data "+
			"against the CSI specification")
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
