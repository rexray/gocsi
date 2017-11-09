package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"os"

	"github.com/thecodeteam/gocsi"
	"github.com/thecodeteam/gocsi/csi"
	"google.golang.org/grpc"
)

var nodeCmds = []*cmd{
	&cmd{
		Name:    "nodepublishvolume",
		Aliases: []string{"mnt", "mount"},
		Action:  nodePublishVolume,
		Flags:   flagsNodePublishVolume,
	},
	&cmd{
		Name:    "nodeunpublishvolume",
		Aliases: []string{"umount", "unmount"},
		Action:  nodeUnpublishVolume,
		Flags:   flagsNodeUnpublishVolume,
	},
	&cmd{
		Name:    "getnodeid",
		Aliases: []string{"id", "getn", "nodeid"},
		Action:  getNodeID,
		Flags:   flagsGetNodeID,
	},
	&cmd{
		Name:    "nodeprobe",
		Aliases: []string{"np", "nprobe"},
		Action:  nodeProbe,
		Flags:   flagsNodeProbe,
	},
	&cmd{
		Name:    "nodegetcapabilities",
		Aliases: []string{"n", "node", "nget"},
		Action:  nodeGetCapabilities,
		Flags:   flagsNodeGetCapabilities,
	},
}

///////////////////////////////////////////////////////////////////////////////
//                            NodePublishVolume                              //
///////////////////////////////////////////////////////////////////////////////
var argsNodePublishVolume struct {
	volumeAT          mapOfStringArg
	publishVolumeInfo mapOfStringArg
	targetPath        string
	fsType            string
	mntFlags          stringSliceArg
	readOnly          bool
	mode              int64
	block             bool
}

func flagsNodePublishVolume(
	ctx context.Context, rpc string) *flag.FlagSet {

	fs := flag.NewFlagSet(rpc, flag.ExitOnError)
	flagsGlobal(fs, "", "")

	fs.Var(
		&argsNodePublishVolume.volumeAT,
		"attribs",
		"The volume attributes.")

	fs.Var(
		&argsNodePublishVolume.publishVolumeInfo,
		"publishVolumeInfo",
		"The published volume info to use.")

	fs.StringVar(
		&argsNodePublishVolume.targetPath,
		"targetPath",
		"",
		"The path to which the volume will be published.")

	fs.BoolVar(
		&argsNodePublishVolume.block,
		"block",
		false,
		"A flag that marks the volume for raw device access")

	fs.Int64Var(
		&argsNodePublishVolume.mode,
		"mode",
		0,
		"The volume access mode")

	fs.StringVar(
		&argsNodePublishVolume.fsType,
		"t",
		"",
		"The file system type")

	fs.Var(
		&argsNodePublishVolume.mntFlags,
		"o",
		"The mount flags")

	fs.BoolVar(
		&argsNodePublishVolume.readOnly,
		"ro",
		false,
		"A flag indicating whether or not to "+
			"publish the volume in read-only mode.")

	fs.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			"usage: %s %s [ARGS...] ID\n",
			appName, rpc)
		fs.PrintDefaults()
	}

	return fs
}

func nodePublishVolume(
	ctx context.Context,
	fs *flag.FlagSet,
	cc *grpc.ClientConn) error {

	var (
		client  csi.NodeClient
		version = args.version

		volumeID   string
		mode       csi.VolumeCapability_AccessMode_Mode
		capability *csi.VolumeCapability

		block      = argsNodePublishVolume.block
		fsType     = argsNodePublishVolume.fsType
		mntFlags   = argsNodePublishVolume.mntFlags.vals
		volumeAT   = argsNodePublishVolume.volumeAT.vals
		pubVolInfo = argsNodePublishVolume.publishVolumeInfo.vals
		targetPath = argsNodePublishVolume.targetPath
		readOnly   = argsNodePublishVolume.readOnly
	)

	// make sure maxEntries doesn't exceed int32
	if max := argsNodePublishVolume.mode; max > maxInt32 {
		return fmt.Errorf("error: max entries > int32: %v", max)
	}
	mode = csi.VolumeCapability_AccessMode_Mode(argsNodePublishVolume.mode)

	if block {
		capability = gocsi.NewBlockCapability(mode)
	} else {
		capability = gocsi.NewMountCapability(mode, fsType, mntFlags)
	}

	// If there are unprocessed tokens then set the first one
	// as the volume ID.
	if fs.NArg() > 0 {
		volumeID = fs.Arg(0)
	}

	// initialize the csi client
	client = csi.NewNodeClient(cc)

	// execute the rpc
	err := gocsi.NodePublishVolume(
		ctx, client, version, volumeID,
		volumeAT, pubVolInfo, targetPath,
		capability, readOnly, userCreds)
	if err != nil {
		return err
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////////
//                           NodeUnpublishVolume                             //
///////////////////////////////////////////////////////////////////////////////
var argsNodeUnpublishVolume struct {
	volumeAT   mapOfStringArg
	targetPath string
}

func flagsNodeUnpublishVolume(
	ctx context.Context, rpc string) *flag.FlagSet {

	fs := flag.NewFlagSet(rpc, flag.ExitOnError)
	flagsGlobal(fs, "", "")

	fs.StringVar(
		&argsNodeUnpublishVolume.targetPath,
		"targetPath",
		"",
		"The path to which the volume is published.")

	fs.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			"usage: %s %s [ARGS...] ID\n",
			appName, rpc)
		fs.PrintDefaults()
	}
	return fs
}

func nodeUnpublishVolume(
	ctx context.Context,
	fs *flag.FlagSet,
	cc *grpc.ClientConn) error {

	var (
		client     csi.NodeClient
		version    = args.version
		volumeID   string
		targetPath = argsNodeUnpublishVolume.targetPath
	)

	// If there are unprocessed tokens then set the first one
	// as the volume ID.
	if fs.NArg() > 0 {
		volumeID = fs.Arg(0)
	}

	// initialize the csi client
	client = csi.NewNodeClient(cc)

	// execute the rpc
	err := gocsi.NodeUnpublishVolume(
		ctx, client, version, volumeID, targetPath, userCreds)
	if err != nil {
		return err
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////////
//                                GetNodeID                                  //
///////////////////////////////////////////////////////////////////////////////
func flagsGetNodeID(
	ctx context.Context, rpc string) *flag.FlagSet {

	fs := flag.NewFlagSet(rpc, flag.ExitOnError)
	flagsGlobal(fs, "", "")

	fs.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			"usage: %s %s [ARGS...]\n",
			appName, rpc)
		fs.PrintDefaults()
	}
	return fs
}

func getNodeID(
	ctx context.Context,
	fs *flag.FlagSet,
	cc *grpc.ClientConn) error {

	var (
		client  csi.NodeClient
		version = args.version
		err     error
		nodeID  string
	)

	// initialize the csi client
	client = csi.NewNodeClient(cc)

	// execute the rpc
	if nodeID, err = gocsi.GetNodeID(ctx, client, version); err != nil {
		return err
	}

	// emit the result
	fmt.Println(nodeID)
	return nil
}

///////////////////////////////////////////////////////////////////////////////
//                                ProbeNode                                  //
///////////////////////////////////////////////////////////////////////////////
func flagsNodeProbe(
	ctx context.Context, rpc string) *flag.FlagSet {

	fs := flag.NewFlagSet(rpc, flag.ExitOnError)
	flagsGlobal(fs, "", "")

	fs.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			"usage: %s %s [ARGS...]\n",
			appName, rpc)
		fs.PrintDefaults()
	}
	return fs
}

func nodeProbe(
	ctx context.Context,
	fs *flag.FlagSet,
	cc *grpc.ClientConn) error {

	// initialize the csi client
	client := csi.NewNodeClient(cc)

	// execute the rpc
	err := gocsi.NodeProbe(ctx, client, args.version)
	if err != nil {
		return err
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////////
//                              NodeGetCapabilities                          //
///////////////////////////////////////////////////////////////////////////////
func flagsNodeGetCapabilities(
	ctx context.Context, rpc string) *flag.FlagSet {

	fs := flag.NewFlagSet(rpc, flag.ExitOnError)
	flagsGlobal(fs, capFormat, "[]*csi.NodeServiceCapability")

	fs.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			"usage: %s %s [ARGS...]\n",
			appName, rpc)
		fs.PrintDefaults()
	}
	return fs
}

func nodeGetCapabilities(
	ctx context.Context,
	fs *flag.FlagSet,
	cc *grpc.ClientConn) error {

	// initialize the csi client
	client := csi.NewNodeClient(cc)

	// execute the rpc
	caps, err := gocsi.NodeGetCapabilities(ctx, client, args.version)
	if err != nil {
		return err
	}

	// create a template for emitting the output
	tpl := template.New("template")
	if tpl, err = tpl.Parse(args.format); err != nil {
		return err
	}
	// emit the results
	for _, c := range caps {
		if err = tpl.Execute(
			os.Stdout, c); err != nil {
			return err
		}
	}

	return nil
}
