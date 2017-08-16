package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/codedellemc/gocsi"
	"github.com/codedellemc/gocsi/csi"
)

const (
	// defaultVersion is the default CSI_VERSION string if none
	// is provided via a CLI argument or environment variable
	defaultVersion = "0.0.0"

	// maxUint32 is the maximum value for a uint32. this is
	// defined as math.MaxUint32, but it's redefined here
	// in order to avoid importing the math package for just
	// a constant value
	maxUint32 = 4294967295
)

var appName = path.Base(os.Args[0])

func main() {

	// the program should have at least two args:
	//
	//     args[0]  path of executable
	//     args[1]  csi rpc
	if len(os.Args) < 2 {
		usage(os.Stderr)
		os.Exit(1)
	}

	// match the name of the rpc or one of its aliases
	rpc := os.Args[1]
	c := func(ccc ...[]*cmd) *cmd {
		for _, cc := range ccc {
			for _, c := range cc {
				if strings.EqualFold(rpc, c.Name) {
					rpc = c.Name
					return c
				}
				for _, a := range c.Aliases {
					if strings.EqualFold(rpc, a) {
						rpc = a
						return c
					}
				}
			}
		}
		return nil
	}(controllerCmds, identityCmds, nodeCmds)

	// assert that a command for the requested rpc was found
	if c == nil {
		fmt.Fprintf(os.Stderr, "error: invalid rpc: %s\n", rpc)
		usage(os.Stderr)
		os.Exit(1)
	}

	if c.Action == nil {
		panic("nil rpc action")
	}
	if c.Flags == nil {
		panic("nil rpc flags")
	}

	ctx := context.Background()

	// parse the command line with the command's flag set
	cflags := c.Flags(ctx, rpc)
	if err := cflags.Parse(os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// assert that the endpoint value is required
	if args.endpoint == "" {
		fmt.Fprintln(os.Stderr, "error: endpoint is required")
		cflags.Usage()
		os.Exit(1)
	}

	// assert that the version is required and valid
	versionRX := regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)
	versionMatch := versionRX.FindStringSubmatch(args.szVersion)
	if len(versionMatch) == 0 {
		fmt.Fprintf(
			os.Stderr,
			"error: invalid version: %s\n",
			args.szVersion)
		os.Exit(1)
	}
	versionMajor, _ := strconv.Atoi(versionMatch[1])
	if versionMajor > maxUint32 {
		fmt.Fprintf(
			os.Stderr, "error: MAJOR > uint32: %v\n", versionMajor)
		os.Exit(1)
	}
	versionMinor, _ := strconv.Atoi(versionMatch[2])
	if versionMinor > maxUint32 {
		fmt.Fprintf(
			os.Stderr, "error: MINOR > uint32: %v\n", versionMinor)
		os.Exit(1)
	}
	versionPatch, _ := strconv.Atoi(versionMatch[3])
	if versionPatch > maxUint32 {
		fmt.Fprintf(
			os.Stderr, "error: PATCH > uint32: %v\n", versionPatch)
		os.Exit(1)
	}
	args.version = &csi.Version{
		Major: uint32(versionMajor),
		Minor: uint32(versionMinor),
		Patch: uint32(versionPatch),
	}

	// initialize a grpc client
	gclient, err := newGrpcClient(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// if a service is specified then add it to the context
	// as gRPC metadata
	if args.service != "" {
		ctx = metadata.NewContext(
			ctx, metadata.Pairs("csi.service", args.service))
	}

	// execute the command
	if err := c.Action(ctx, cflags, gclient); err != nil {
		fmt.Fprintln(os.Stderr, err)
		if _, ok := err.(*errUsage); ok {
			cflags.Usage()
		}
		os.Exit(1)
	}
}

///////////////////////////////////////////////////////////////////////////////
//                            Default Formats                                //
///////////////////////////////////////////////////////////////////////////////

// mapSzOfSzFormat is the default Go template format for
// emitting a map[string]string
const mapSzOfSzFormat = `{{range $k, $v := .}}` +
	`{{printf "%s=%s\t" $k $v}}{{end}}{{"\n"}}`

// volumeInfoFormat is the default Go template format for
// emitting a *csi.VolumeInfo
const volumeInfoFormat = `{{with .GetId}}{{range $k, $v := .GetValues}}` +
	`{{printf "%s=%s\t" $k $v}}{{end}}{{end}}{{"\n"}}`

// versionFormat is the default Go template format for emitting a *csi.Version
const versionFormat = `{{.GetMajor}}.{{.GetMinor}}.{{.GetPatch}}`

// pluginInfoFormat is the default Go template format for
// emitting a *csi.GetPluginInfoResponse_Result
const pluginInfoFormat = `{{.Name}}{{print "\t"}}{{.VendorVersion}}{{print "\t"}}` +
	`{{with .GetManifest}}{{range $k, $v := .}}` +
	`{{printf "%s=%s\t" $k $v}}{{end}}{{end}}{{"\n"}}`

const ctrlCapFormat = `{{.GetType}}{{"\n"}}`

///////////////////////////////////////////////////////////////////////////////
//                                Commands                                   //
///////////////////////////////////////////////////////////////////////////////
type errUsage struct {
	msg string
}

func (e *errUsage) Error() string {
	return e.msg
}

type cmd struct {
	Name    string
	Aliases []string
	Action  func(context.Context, *flag.FlagSet, *grpc.ClientConn) error
	Flags   func(context.Context, string) *flag.FlagSet
}

var controllerCmds = []*cmd{
	&cmd{
		Name:    "createvolume",
		Aliases: []string{"new", "create"},
		Action:  createVolume,
		Flags:   flagsCreateVolume,
	},
	&cmd{
		Name:    "deletevolume",
		Aliases: []string{"d", "rm", "del"},
		Action:  nil,
		Flags:   nil,
	},
	&cmd{
		Name:    "controllerpublishvolume",
		Aliases: []string{"att", "attach"},
		Action:  controllerPublishVolume,
		Flags:   flagsControllerPublishVolume,
	},
	&cmd{
		Name:    "controllerunpublishvolume",
		Aliases: []string{"det", "detach"},
		Action:  controllerUnpublishVolume,
		Flags:   flagsControllerUnpublishVolume,
	},
	&cmd{
		Name:    "validatevolumecapabilities",
		Aliases: []string{"v", "validate"},
		Action:  nil,
		Flags:   nil,
	},
	&cmd{
		Name:    "listvolumes",
		Aliases: []string{"l", "ls", "list"},
		Action:  listVolumes,
		Flags:   flagsListVolumes,
	},
	&cmd{
		Name:    "getcapacity",
		Aliases: []string{"getc", "capacity"},
		Action:  nil,
		Flags:   nil,
	},
	&cmd{
		Name:    "controllergetcapabilities",
		Aliases: []string{"cget"},
		Action:  controllerGetCapabilities,
		Flags:   flagsControllerGetCapabilities,
	},
}

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
		Action:  nil,
		Flags:   nil,
	},
	&cmd{
		Name:    "nodegetcapabilities",
		Aliases: []string{"n", "node"},
		Action:  nil,
		Flags:   nil,
	},
}

///////////////////////////////////////////////////////////////////////////////
//                                Usage                                      //
///////////////////////////////////////////////////////////////////////////////
func usage(w io.Writer) {
	const h = `usage: {{.Name}} RPC [ARGS...]{{range $Name, $Cmds := .Categories}}

       {{$Name}} RPCs{{range $Cmds}}
         {{.Name}}{{if .Aliases}} ({{join .Aliases ", "}}){{end}}{{end}}{{end}}

Use the -? flag with an RPC for additional help.
`
	f := template.FuncMap{"join": strings.Join}
	t := template.Must(template.New(appName).Funcs(f).Parse(h))
	d := struct {
		Name       string
		Categories map[string][]*cmd
	}{
		appName,
		map[string][]*cmd{
			"CONTROLLER": controllerCmds,
			"IDENTITY":   identityCmds,
			"NODE":       nodeCmds,
		},
	}
	t.Execute(w, d)
}

///////////////////////////////////////////////////////////////////////////////
//                               Global Flags                                //
///////////////////////////////////////////////////////////////////////////////
var args struct {
	service   string
	endpoint  string
	format    string
	help      bool
	insecure  bool
	szVersion string
	version   *csi.Version
}

func flagsGlobal(
	fs *flag.FlagSet,
	formatDefault, formatObjectType string) {

	fs.StringVar(
		&args.endpoint,
		"endpoint",
		os.Getenv("CSI_ENDPOINT"),
		"The endpoint address")

	fs.StringVar(
		&args.service,
		"service",
		"",
		"The name of the CSD service to use.")

	version := defaultVersion
	if v := os.Getenv("CSI_VERSION"); v != "" {
		version = v
	}
	fs.StringVar(
		&args.szVersion,
		"version",
		version,
		"The API version string")

	insecure := true
	if v := os.Getenv("CSI_INSECURE"); v != "" {
		insecure, _ = strconv.ParseBool(v)
	}
	fs.BoolVar(
		&args.insecure,
		"insecure",
		insecure,
		"Disables transport security")

	fmtMsg := &bytes.Buffer{}
	fmt.Fprint(fmtMsg, "The Go template used to print an object.")
	if formatObjectType != "" {
		fmt.Fprintf(fmtMsg, " This command emits a %s.", formatObjectType)
	}
	fs.StringVar(
		&args.format,
		"format",
		formatDefault,
		fmtMsg.String())
}

///////////////////////////////////////////////////////////////////////////////
//                              CreateVolume                                 //
///////////////////////////////////////////////////////////////////////////////
var argsCreateVolume struct {
	reqBytes uint64
	limBytes uint64
	fsType   string
	mntFlags stringSliceArg
	params   mapOfStringArg
}

func flagsCreateVolume(ctx context.Context, rpc string) *flag.FlagSet {
	fs := flag.NewFlagSet(rpc, flag.ExitOnError)
	flagsGlobal(fs, volumeInfoFormat, "*csi.VolumeInfo")

	fs.Uint64Var(
		&argsCreateVolume.reqBytes,
		"requiredBytes",
		0,
		"The minimum volume size in bytes")

	fs.Uint64Var(
		&argsCreateVolume.limBytes,
		"limitBytes",
		0,
		"The maximum volume size in bytes")

	fs.StringVar(
		&argsCreateVolume.fsType,
		"t",
		"",
		"The file system type")

	fs.Var(
		&argsCreateVolume.mntFlags,
		"o",
		"The mount flags")

	fs.Var(
		&argsCreateVolume.params,
		"params",
		"Additional RPC parameters")

	fs.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			"usage: %s %s [ARGS...] NAME\n",
			appName, rpc)
		fs.PrintDefaults()
	}

	return fs
}

func createVolume(
	ctx context.Context,
	fs *flag.FlagSet,
	cc *grpc.ClientConn) error {

	var (
		client csi.ControllerClient
		err    error
		tpl    *template.Template

		name     = fs.Arg(0)
		reqBytes = argsCreateVolume.reqBytes
		limBytes = argsCreateVolume.limBytes
		fsType   = argsCreateVolume.fsType
		mntFlags = argsCreateVolume.mntFlags.vals
		params   = argsCreateVolume.params.vals

		format  = args.format
		version = args.version
	)

	// create a template for emitting the output
	tpl = template.New("template")
	if tpl, err = tpl.Parse(format); err != nil {
		return err
	}

	// initialize the csi client
	client = csi.NewControllerClient(cc)

	// execute the rpc
	result, err := gocsi.CreateVolume(
		ctx, client, version, name,
		reqBytes, limBytes,
		fsType, mntFlags, params)
	if err != nil {
		return err
	}

	// emit the result
	if err = tpl.Execute(os.Stdout, result); err != nil {
		return err
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////////
//                          ControllerPublishVolume                          //
///////////////////////////////////////////////////////////////////////////////
var argsControllerPublishVolume struct {
	volumeMD mapOfStringArg
	nodeID   mapOfStringArg
	readOnly bool
}

func flagsControllerPublishVolume(
	ctx context.Context, rpc string) *flag.FlagSet {

	fs := flag.NewFlagSet(rpc, flag.ExitOnError)
	flagsGlobal(fs, mapSzOfSzFormat, "map[string]string")

	fs.Var(
		&argsControllerPublishVolume.volumeMD,
		"metadata",
		"The metadata of the volume to be used on a node.")

	fs.Var(
		&argsControllerPublishVolume.nodeID,
		"nodeID",
		"The ID of the node to which the volume should be published.")

	fs.BoolVar(
		&argsControllerPublishVolume.readOnly,
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

func controllerPublishVolume(
	ctx context.Context,
	fs *flag.FlagSet,
	cc *grpc.ClientConn) error {

	if fs.NArg() == 0 {
		return &errUsage{"missing volume ID"}
	}

	var (
		client csi.ControllerClient
		err    error
		tpl    *template.Template

		volumeMD *csi.VolumeMetadata
		nodeID   *csi.NodeID

		volumeID = &csi.VolumeID{Values: map[string]string{}}
		readOnly = argsControllerPublishVolume.readOnly

		format  = args.format
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
	if v := argsControllerPublishVolume.volumeMD.vals; len(v) > 0 {
		volumeMD = &csi.VolumeMetadata{Values: v}
	}

	// check for a node ID
	if v := argsControllerPublishVolume.nodeID.vals; len(v) > 0 {
		nodeID = &csi.NodeID{Values: v}
	}

	// create a template for emitting the output
	tpl = template.New("template")
	if tpl, err = tpl.Parse(format); err != nil {
		return err
	}

	// initialize the csi client
	client = csi.NewControllerClient(cc)

	// execute the rpc
	result, err := gocsi.ControllerPublishVolume(
		ctx, client, version, volumeID,
		volumeMD, nodeID, readOnly)
	if err != nil {
		return err
	}

	// emit the result
	if err = tpl.Execute(os.Stdout, result.GetValues()); err != nil {
		return err
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////////
//                        ControllerUnpublishVolume                          //
///////////////////////////////////////////////////////////////////////////////
var argsControllerUnpublishVolume struct {
	volumeMD mapOfStringArg
	nodeID   mapOfStringArg
}

func flagsControllerUnpublishVolume(
	ctx context.Context, rpc string) *flag.FlagSet {

	fs := flag.NewFlagSet(rpc, flag.ExitOnError)
	flagsGlobal(fs, "", "")

	fs.Var(
		&argsControllerUnpublishVolume.volumeMD,
		"metadata",
		"The metadata of the volume.")

	fs.Var(
		&argsControllerUnpublishVolume.nodeID,
		"nodeID",
		"The ID of the node on which the volume is published.")

	fs.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			"usage: %s %s [ARGS...] ID_KEY[=ID_VAL] [ID_KEY[=ID_VAL]...]\n",
			appName, rpc)
		fs.PrintDefaults()
	}

	return fs
}

func controllerUnpublishVolume(
	ctx context.Context,
	fs *flag.FlagSet,
	cc *grpc.ClientConn) error {

	if fs.NArg() == 0 {
		return &errUsage{"missing volume ID"}
	}

	var (
		client csi.ControllerClient

		volumeMD *csi.VolumeMetadata
		nodeID   *csi.NodeID

		volumeID = &csi.VolumeID{Values: map[string]string{}}

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
	if v := argsControllerUnpublishVolume.volumeMD.vals; len(v) > 0 {
		volumeMD = &csi.VolumeMetadata{Values: v}
	}

	// check for a node ID
	if v := argsControllerUnpublishVolume.nodeID.vals; len(v) > 0 {
		nodeID = &csi.NodeID{Values: v}
	}

	// initialize the csi client
	client = csi.NewControllerClient(cc)

	// execute the rpc
	err := gocsi.ControllerUnpublishVolume(
		ctx, client, version, volumeID, volumeMD, nodeID)
	if err != nil {
		return err
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////////
//                              ListVolumes                                  //
///////////////////////////////////////////////////////////////////////////////
var argsListVolumes struct {
	startingToken string
	maxEntries    uint64
	paging        bool
}

func flagsListVolumes(ctx context.Context, rpc string) *flag.FlagSet {
	fs := flag.NewFlagSet(rpc, flag.ExitOnError)
	flagsGlobal(fs, volumeInfoFormat, "*csi.VolumeInfo")

	fs.StringVar(
		&argsListVolumes.startingToken,
		"startingToken",
		os.Getenv("CSI_STARTING_TOKEN"),
		"A token to specify where to start paginating")

	var evMaxEntries uint64
	if v := os.Getenv("CSI_MAX_ENTRIES"); v != "" {
		i, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"error: max entries not uint32: %v\n",
				err)
		}
		evMaxEntries = i
	}
	fs.Uint64Var(
		&argsListVolumes.maxEntries,
		"maxEntries",
		evMaxEntries,
		"The maximum number of entries to return")

	fs.BoolVar(
		&argsListVolumes.paging,
		"paging",
		false,
		"Enables automatic paging")

	fs.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			"usage: %s %s [ARGS...]\n",
			appName, rpc)
		fs.PrintDefaults()
	}

	return fs
}

func listVolumes(
	ctx context.Context,
	fs *flag.FlagSet,
	cc *grpc.ClientConn) error {

	var (
		client     csi.ControllerClient
		err        error
		maxEntries uint32
		tpl        *template.Template
		wg         sync.WaitGroup

		chdone        = make(chan int)
		cherrs        = make(chan error)
		format        = args.format
		startingToken = argsListVolumes.startingToken
		version       = args.version
	)

	// make sure maxEntries doesn't exceed uin32
	if max := argsListVolumes.maxEntries; max > maxUint32 {
		return fmt.Errorf("error: max entries > uin32: %v", max)
	}
	maxEntries = uint32(argsListVolumes.maxEntries)

	// create a template for emitting the output
	tpl = template.New("template")
	if tpl, err = tpl.Parse(format); err != nil {
		return err
	}

	// initialize the csi client
	client = csi.NewControllerClient(cc)

	// the two channels chdone and cherrs are used to
	// track the status of the goroutines as well as
	// the presence of any errors that need to be
	// returned from this function
	wg.Add(1)
	go func() {
		wg.Wait()
		close(chdone)
	}()

	go func() {
		tok := startingToken
		for {
			vols, next, err := gocsi.ListVolumes(
				ctx,
				client,
				version,
				maxEntries,
				tok)
			if err != nil {
				cherrs <- err
				return
			}
			wg.Add(1)
			go func(vols []*csi.VolumeInfo) {
				for _, v := range vols {
					if err := tpl.Execute(os.Stdout, v); err != nil {
						cherrs <- err
						return
					}
				}
				wg.Done()
			}(vols)
			if !argsListVolumes.paging || next == "" {
				break
			}
			tok = next
		}
		wg.Done()
	}()

	select {
	case <-chdone:
	case err := <-cherrs:
		if err != nil {
			return err
		}
	}

	return nil
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
var argsGetNodeID struct {
	volumeMD   mapOfStringArg
	targetPath string
}

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

// newGrpcClient should not be invoked until after flags are parsed
func newGrpcClient(ctx context.Context) (*grpc.ClientConn, error) {
	// the grpc dialer *assumes* tcp, which is silly. this custom
	// dialer parses the network protocol from a fully-formed golang
	// network string and defers the dialing to net.DialTimeout
	endpoint := args.endpoint
	dialOpts := []grpc.DialOption{
		grpc.WithDialer(
			func(target string, timeout time.Duration) (net.Conn, error) {
				proto, addr, err := gocsi.ParseProtoAddr(target)
				if err != nil {
					return nil, err
				}
				return net.DialTimeout(proto, addr, timeout)
			}),
	}
	if args.insecure {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}
	return grpc.DialContext(ctx, endpoint, dialOpts...)
}

// stringSliceArg is used for parsing a csv arg into a string slice
type stringSliceArg struct {
	szVal string
	vals  []string
}

func (s *stringSliceArg) String() string {
	return s.szVal
}

func (s *stringSliceArg) Set(val string) error {
	s.vals = append(s.vals, strings.Split(val, ",")...)
	return nil
}

// mapOfStringArg is used for parsing a csv, key=value arg into
// a map[string]string
type mapOfStringArg struct {
	szVal string
	vals  map[string]string
}

func (s *mapOfStringArg) String() string {
	return s.szVal
}

func (s *mapOfStringArg) Set(val string) error {
	if s.vals == nil {
		s.vals = map[string]string{}
	}
	vals := strings.Split(val, ",")
	for _, v := range vals {
		vp := strings.SplitN(v, "=", 2)
		switch len(vp) {
		case 1:
			s.vals[vp[0]] = ""
		case 2:
			s.vals[vp[0]] = vp[1]
		}
	}
	return nil
}

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

	// initialize the csi client
	client := csi.NewIdentityClient(cc)

	// execute the rpc
	versions, err := gocsi.GetSupportedVersions(ctx, client)
	if err != nil {
		return err
	}

	// create a template for emitting the output
	tpl := template.New("template")
	if tpl, err = tpl.Parse(args.format); err != nil {
		return err
	}
	// emit the result
	for _, v := range versions {
		if err = tpl.Execute(
			os.Stdout, v); err != nil {
			return err
		}
	}

	return nil
}

func flagsGetPluginInfo(
	ctx context.Context, rpc string) *flag.FlagSet {

	fs := flag.NewFlagSet(rpc, flag.ExitOnError)
	flagsGlobal(fs, pluginInfoFormat, "*csi.GetPluginInfoResponse_Result")

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

	// initialize the csi client
	client := csi.NewIdentityClient(cc)

	// execute the rpc
	info, err := gocsi.GetPluginInfo(ctx, client, args.version)
	if err != nil {
		return err
	}

	// create a template for emitting the output
	tpl := template.New("template")
	if tpl, err = tpl.Parse(args.format); err != nil {
		return err
	}
	// emit the result
	if err = tpl.Execute(
		os.Stdout, info); err != nil {
		return err
	}

	return nil
}

func flagsControllerGetCapabilities(
	ctx context.Context, rpc string) *flag.FlagSet {

	fs := flag.NewFlagSet(rpc, flag.ExitOnError)
	flagsGlobal(fs, ctrlCapFormat, "*csi.ControllerGetCapabilitiesResponse_Result")

	fs.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			"usage: %s %s [ARGS...]\n",
			appName, rpc)
		fs.PrintDefaults()
	}
	return fs
}

func controllerGetCapabilities(
	ctx context.Context,
	fs *flag.FlagSet,
	cc *grpc.ClientConn) error {

	// initialize the csi client
	client := csi.NewControllerClient(cc)

	// execute the rpc
	caps, err := gocsi.ControllerGetCapabilities(ctx, client, args.version)
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
