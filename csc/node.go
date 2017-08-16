package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/codedellemc/gocsi"
	"github.com/codedellemc/gocsi/csi"
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
		Name:    "probenode",
		Aliases: []string{"p", "probe"},
		Action:  probeNode,
		Flags:   flagsProbeNode,
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
	volumeMD          mapOfStringArg
	publishVolumeInfo mapOfStringArg
	targetPath        string
	fsType            string
	mntFlags          stringSliceArg
	readOnly          bool
}

func flagsNodePublishVolume(
	ctx context.Context, rpc string) *flag.FlagSet {

	fs := flag.NewFlagSet(rpc, flag.ExitOnError)
	flagsGlobal(fs, "", "")

	fs.Var(
		&argsNodePublishVolume.volumeMD,
		"metadata",
		"The metadata of the volume to be used on a node.")

	fs.Var(
		&argsNodePublishVolume.publishVolumeInfo,
		"publishVolumeInfo",
		"The published volume info to use.")

	fs.StringVar(
		&argsNodePublishVolume.targetPath,
		"targetPath",
		"",
		"The path to which the volume will be published.")

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
			"usage: %s %s [ARGS...] ID_KEY[=ID_VAL] [ID_KEY[=ID_VAL]...]\n",
			appName, rpc)
		fs.PrintDefaults()
	}

	return fs
}

func nodePublishVolume(
	ctx context.Context,
	fs *flag.FlagSet,
	cc *grpc.ClientConn) error {

	if fs.NArg() == 0 {
		return &errUsage{"missing volume ID"}
	}
	if argsNodePublishVolume.targetPath == "" {
		return &errUsage{"missing targetPath"}
	}
	if argsNodePublishVolume.fsType == "" {
		return &errUsage{"missing fsType"}
	}
	if len(argsNodePublishVolume.mntFlags.vals) == 0 {
		return &errUsage{"missing mount flags (-o)"}
	}

	var (
		client csi.NodeClient

		volumeMD   *csi.VolumeMetadata
		pubVolInfo *csi.PublishVolumeInfo

		capability = &csi.VolumeCapability{
			Value: &csi.VolumeCapability_Mount{
				Mount: &csi.VolumeCapability_MountVolume{
					FsType:     argsNodePublishVolume.fsType,
					MountFlags: argsNodePublishVolume.mntFlags.vals,
				},
			},
		}
		volumeID   = &csi.VolumeID{Values: map[string]string{}}
		targetPath = argsNodePublishVolume.targetPath
		readOnly   = argsNodePublishVolume.readOnly

		version = args.version
	)

	// parse the volume ID into a map
	for x := 0; x < fs.NArg(); x++ {
		a := fs.Arg(x)
		kv := strings.SplitN(a, "=", 2)
		switch len(kv) {
		case 1:
			volumeID.Values[kv[0]] = ""
		case 2:
			volumeID.Values[kv[0]] = kv[1]
		}
	}

	// check for volume metadata
	if v := argsNodePublishVolume.volumeMD.vals; len(v) > 0 {
		volumeMD = &csi.VolumeMetadata{Values: v}
	}

	// check for publish volume info
	if v := argsNodePublishVolume.publishVolumeInfo.vals; len(v) > 0 {
		pubVolInfo = &csi.PublishVolumeInfo{Values: v}
	}

	// initialize the csi client
	client = csi.NewNodeClient(cc)

	// execute the rpc
	err := gocsi.NodePublishVolume(
		ctx, client, version, volumeID,
		volumeMD, pubVolInfo, targetPath,
		capability, readOnly)
	if err != nil {
		return err
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////////
//                           NodeUnpublishVolume                             //
///////////////////////////////////////////////////////////////////////////////
var argsNodeUnpublishVolume struct {
	volumeMD   mapOfStringArg
	targetPath string
}

func flagsNodeUnpublishVolume(
	ctx context.Context, rpc string) *flag.FlagSet {

	fs := flag.NewFlagSet(rpc, flag.ExitOnError)
	flagsGlobal(fs, "", "")

	fs.Var(
		&argsNodeUnpublishVolume.volumeMD,
		"metadata",
		"The metadata of the volume to be used on a node.")

	fs.StringVar(
		&argsNodeUnpublishVolume.targetPath,
		"targetPath",
		"",
		"The path to which the volume is published.")

	fs.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			"usage: %s %s [ARGS...] ID_KEY[=ID_VAL] [ID_KEY[=ID_VAL]...]\n",
			appName, rpc)
		fs.PrintDefaults()
	}
	return fs
}

func nodeUnpublishVolume(
	ctx context.Context,
	fs *flag.FlagSet,
	cc *grpc.ClientConn) error {

	if fs.NArg() == 0 {
		return &errUsage{"missing volume ID"}
	}
	if argsNodeUnpublishVolume.targetPath == "" {
		return &errUsage{"missing targetPath"}
	}

	var (
		client csi.NodeClient

		volumeMD *csi.VolumeMetadata

		volumeID   = &csi.VolumeID{Values: map[string]string{}}
		targetPath = argsNodeUnpublishVolume.targetPath

		version = args.version
	)

	// parse the volume ID into a map
	for x := 0; x < fs.NArg(); x++ {
		a := fs.Arg(x)
		kv := strings.SplitN(a, "=", 2)
		switch len(kv) {
		case 1:
			volumeID.Values[kv[0]] = ""
		case 2:
			volumeID.Values[kv[0]] = kv[1]
		}
	}

	// check for volume metadata
	if v := argsNodeUnpublishVolume.volumeMD.vals; len(v) > 0 {
		volumeMD = &csi.VolumeMetadata{Values: v}
	}

	// initialize the csi client
	client = csi.NewNodeClient(cc)

	// execute the rpc
	err := gocsi.NodeUnpublishVolume(
		ctx, client, version, volumeID,
		volumeMD, targetPath)
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
	flagsGlobal(fs, mapSzOfSzFormat, "map[string]string")

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
		err     error
		client  csi.NodeClient
		tpl     *template.Template
		nodeID  *csi.NodeID
		version = args.version
		format  = args.format
	)

	// create a template for emitting the output
	tpl = template.New("template")
	if tpl, err = tpl.Parse(format); err != nil {
		return err
	}

	// initialize the csi client
	client = csi.NewNodeClient(cc)

	// execute the rpc
	if nodeID, err = gocsi.GetNodeID(ctx, client, version); err != nil {
		return err
	}

	// emit the result
	if err = tpl.Execute(
		os.Stdout, nodeID.GetValues()); err != nil {
		return err
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////////
//                                ProbeNode                                  //
///////////////////////////////////////////////////////////////////////////////
func flagsProbeNode(
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

func probeNode(
	ctx context.Context,
	fs *flag.FlagSet,
	cc *grpc.ClientConn) error {

	// initialize the csi client
	client := csi.NewNodeClient(cc)

	// execute the rpc
	err := gocsi.ProbeNode(ctx, client, args.version)
	if err != nil {
		return err
	}

	fmt.Println("Success")

	return nil
}

///////////////////////////////////////////////////////////////////////////////
//                              NodeGetCapabilities                          //
///////////////////////////////////////////////////////////////////////////////
func flagsNodeGetCapabilities(
	ctx context.Context, rpc string) *flag.FlagSet {

	fs := flag.NewFlagSet(rpc, flag.ExitOnError)
	flagsGlobal(fs, capFormat, "*csi.nodeGetCapabilitiesResponse_Result")

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
