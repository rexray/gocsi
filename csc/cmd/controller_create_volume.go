package cmd

import (
	"context"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var createVolume struct {
	reqBytes uint64
	limBytes uint64
	caps     volumeCapabilitySliceArg
	params   mapOfStringArg
	reqCreds bool
}

var createVolumeCmd = &cobra.Command{
	Use:     "create-volume",
	Aliases: []string{"n", "c", "new", "create"},
	Short:   `invokes the rpc "CreateVolume"`,
	Example: `
CREATING MULTIPLE VOLUMES
        The following example illustrates how to create two volumes with the
        same characteristics at the same time:

            csc controller new --endpoint /csi/server.sock
                               --cap 1,block \
                               --cap MULTI_NODE_MULTI_WRITER,mount,xfs,uid=500 \
                               --params region=us,zone=texas
                               --params disabled=false
                               MyNewVolume1 MyNewVolume2
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		req := csi.CreateVolumeRequest{
			Version:            &root.version.Version,
			VolumeCapabilities: createVolume.caps.data,
			Parameters:         createVolume.params.data,
			UserCredentials:    root.userCreds,
		}

		if createVolume.reqBytes > 0 || createVolume.limBytes > 0 {
			req.CapacityRange = &csi.CapacityRange{}
			if v := createVolume.reqBytes; v > 0 {
				req.CapacityRange.RequiredBytes = v
			}
			if v := createVolume.limBytes; v > 0 {
				req.CapacityRange.LimitBytes = v
			}
		}

		for i := range args {
			ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
			defer cancel()

			// Set the volume name for the current request.
			req.Name = args[i]

			log.WithField("request", req).Debug("creating volume")
			rep, err := controller.client.CreateVolume(ctx, &req)
			if err != nil {
				return err
			}
			if err := root.tpl.Execute(os.Stdout, rep.VolumeInfo); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	controllerCmd.AddCommand(createVolumeCmd)

	createVolumeCmd.Flags().Uint64Var(
		&createVolume.reqBytes,
		"req-bytes",
		0,
		"The required size of the volume in bytes")

	createVolumeCmd.Flags().Uint64Var(
		&createVolume.limBytes,
		"lim-bytes",
		0,
		"The limit to the size of the volume in bytes")

	createVolumeCmd.Flags().Var(
		&createVolume.params,
		"params",
		`One or more key/value pairs may be specified to send with
        the request as its Parameters field:

            --params key1=val1,key2=val2 --params=key3=val3`)

	flagVolumeCapabilities(createVolumeCmd.Flags(), &createVolume.caps)

	flagWithRequiresCreds(
		createVolumeCmd.Flags(),
		&root.withRequiresCreds,
		"")

	flagWithRequiresAttribs(
		createVolumeCmd.Flags(),
		&root.withRequiresVolumeAttributes,
		"")
}
