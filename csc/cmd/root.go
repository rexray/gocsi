package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"text/template"
	"time"

	"github.com/dell/gocsi/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var debug, _ = strconv.ParseBool(os.Getenv("X_CSI_DEBUG"))

var root struct {
	ctx     context.Context
	client  *grpc.ClientConn
	tpl     *template.Template
	secrets map[string]string

	genMarkdown bool
	logLevel    logLevelArg
	format      string
	endpoint    string
	insecure    bool
	timeout     time.Duration
	metadata    mapOfStringArg

	withReqLogging bool
	withRepLogging bool

	withSpecValidator      bool
	withRequiresCreds      bool
	withRequiresVolContext bool
	withRequiresPubContext bool
}

var (
	activeArgs []string
	activeCmd  *cobra.Command
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "csc",
	Short: "a command line container storage interface (CSI) client",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

		// Enable debug level logging and request and response logging
		// if the environment variable that controls deubg mode is set
		// to a truthy value.
		if debug {
			root.logLevel.Set(log.DebugLevel.String())
			root.withReqLogging = true
			root.withReqLogging = true
		}

		// Set the log level.
		lvl, _ := root.logLevel.Val()
		log.SetLevel(lvl)

		if debug {
			log.Warn("debug mode enabled")
		}

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
			case listSnapshotsCmd.Name():
				if listSnapshots.paging {
					root.format = snapshotInfoFormat
				} else {
					root.format = listSnapshotsFormat
				}
			case createSnapshotCmd.Name():
				root.format = snapshotInfoFormat
			case createVolumeCmd.Name():
				root.format = volumeInfoFormat
			case pluginInfoCmd.Name():
				root.format = pluginInfoFormat
			case pluginCapsCmd.Name():
				root.format = pluginCapsFormat
			case probeCmd.Name():
				root.format = probeFormat
			case nodeGetVolumeStatsCmd.Name():
				root.format = statsFormat
			case nodeGetInfoCmd.Name():
				root.format = nodeInfoFormat
			}
		}
		if root.format != "" {
			tpl, err := template.New("t").Funcs(template.FuncMap{
				"isa": func(o interface{}, t string) bool {
					return fmt.Sprintf("%T", o) == t
				},
			}).Parse(root.format)
			if err != nil {
				return err
			}
			root.tpl = tpl
		}

		// Parse the credentials if they exist.
		root.secrets = utils.ParseMap(os.Getenv("X_CSI_SECRETS"))

		// Create the gRPC client connection.
		opts := []grpc.DialOption{
			grpc.WithDialer(
				func(string, time.Duration) (net.Conn, error) {
					proto, addr, err := utils.ParseProtoAddr(root.endpoint)
					log.WithFields(map[string]interface{}{
						"proto":   proto,
						"addr":    addr,
						"timeout": root.timeout,
					}).Debug("parsed endpoint info")
					if err != nil {
						return nil, err
					}
					return net.DialTimeout(proto, addr, root.timeout)
				}),
		}

		// Disable TLS if specified.
		if root.insecure {
			opts = append(opts, grpc.WithInsecure())
		}

		// Add interceptors to the client if any are configured.
		if o := getClientInterceptorsDialOpt(); o != nil {
			opts = append(opts, o)
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
		exitCode := 1
		if stat, ok := status.FromError(err); ok {
			exitCode = int(stat.Code())
			fmt.Fprintln(os.Stderr, stat.Message())
		} else {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
		fmt.Fprintf(os.Stderr, "\nPlease use -h,--help for more information\n")
		os.Exit(exitCode)
	}
}

func init() {
	setHelpAndUsage(RootCmd)

	flagLogLevel(
		RootCmd.PersistentFlags(),
		&root.logLevel,
		"warn")

	flagEndpoint(
		RootCmd.PersistentFlags(),
		&root.endpoint,
		os.Getenv("CSI_ENDPOINT"))

	flagTimeout(
		RootCmd.PersistentFlags(),
		&root.timeout,
		"1m")

	flagWithRequestLogging(
		RootCmd.PersistentFlags(),
		&root.withReqLogging,
		"false")

	flagWithResponseLogging(
		RootCmd.PersistentFlags(),
		&root.withRepLogging,
		"false")

	flagWithSpecValidation(
		RootCmd.PersistentFlags(),
		&root.withSpecValidator,
		"false")

	RootCmd.PersistentFlags().BoolVarP(
		&root.insecure,
		"insecure",
		"i",
		true,
		`Disables transport security for the client via the gRPC dial option
        WithInsecure (https://goo.gl/Y95SfW)`)

	RootCmd.PersistentFlags().VarP(
		&root.metadata,
		"metadata",
		"m",
		`Sets one or more key/value pairs to use as gRPC metadata sent with all
        RPCs. gRPC metadata is similar to HTTP headers. For example:

            --metadata key1=val1 --m key2=val2,key3=val3

            -m key1=val1,key2=val2 --metadata key3=val3

        Read more on gRPC metadata at https://goo.gl/iTci67`)
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
