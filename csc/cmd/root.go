package cmd

import (
	"context"
	"fmt"
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
	ctx    context.Context
	client *grpc.ClientConn
	tpl    *template.Template

	logLevel string
	format   string
	endpoint string
	insecure bool
	timeout  time.Duration
	version  csiVersionArg
	metadata mapOfStringArg
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
			}
		}
		if root.format != "" {
			tpl, err := template.New("t").Parse(root.format)
			if err != nil {
				return err
			}
			root.tpl = tpl
		}

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
		if root.insecure {
			opts = append(opts, grpc.WithInsecure())
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
}
