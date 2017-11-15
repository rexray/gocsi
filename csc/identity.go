package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"os"

	"github.com/thecodeteam/gocsi/csi"
	"google.golang.org/grpc"
)

var identityCmds = []*cmd{
	&cmd{
		Name:    "getsupportedversions",
		Aliases: []string{"gets"},
		Action:  getSupportedVersions,
		Flags:   flagsGetSupportedVersions,
	},
	&cmd{
		Name:    "getplugininfo",
		Aliases: []string{"getp"},
		Action:  getPluginInfo,
		Flags:   flagsGetPluginInfo,
	},
}

///////////////////////////////////////////////////////////////////////////////
//                          GetSupportedVersions                             //
///////////////////////////////////////////////////////////////////////////////
func flagsGetSupportedVersions(
	ctx context.Context, rpc string) *flag.FlagSet {

	fs := flag.NewFlagSet(rpc, flag.ExitOnError)
	flagsGlobal(fs, versionFormat, "*csi.Version")

	fs.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			"usage: %s %s [ARGS...]\n",
			appName, rpc)
		fs.PrintDefaults()
	}
	return fs
}

func getSupportedVersions(
	ctx context.Context,
	fs *flag.FlagSet,
	cc *grpc.ClientConn) error {

	// Create a template to emit the result.
	tpl, err := template.New("template").Parse(args.format)
	if err != nil {
		return err
	}

	// Execute the RPC.
	client := csi.NewIdentityClient(cc)
	res, err := client.GetSupportedVersions(
		ctx, &csi.GetSupportedVersionsRequest{})
	if err != nil {
		return err
	}

	// Emit the result.
	for _, v := range res.SupportedVersions {
		if err = tpl.Execute(os.Stdout, v); err != nil {
			return err
		}
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////////
//                          GetPluginInfo                                    //
///////////////////////////////////////////////////////////////////////////////
func flagsGetPluginInfo(
	ctx context.Context, rpc string) *flag.FlagSet {

	fs := flag.NewFlagSet(rpc, flag.ExitOnError)
	flagsGlobal(fs, pluginInfoFormat, "*csi.GetPluginInfoResponse")

	fs.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			"usage: %s %s [ARGS...]\n",
			appName, rpc)
		fs.PrintDefaults()
	}
	return fs
}

func getPluginInfo(
	ctx context.Context,
	fs *flag.FlagSet,
	cc *grpc.ClientConn) error {

	// Create a template to emit the result.
	tpl, err := template.New("template").Parse(args.format)
	if err != nil {
		return err
	}

	// Execute the RPC.
	client := csi.NewIdentityClient(cc)
	res, err := client.GetPluginInfo(ctx, &csi.GetPluginInfoRequest{
		Version: args.version,
	})
	if err != nil {
		return err
	}

	// Emit the result.
	if err = tpl.Execute(os.Stdout, res); err != nil {
		return err
	}

	return nil
}
