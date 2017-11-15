package cmd

import (
	"context"
	"errors"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/thecodeteam/gocsi/csi"
)

var createVolume struct {
	reqBytes uint64
	limBytes uint64
	caps     volumeCapabilitySliceArg
	params   mapOfStringArg
	reqCreds bool
}

var createVolumeCmd = &cobra.Command{
	Use:     "createvolume",
	Aliases: []string{"c", "n", "new", "create"},
	Short:   `invokes the rpc "CreateVolume"`,
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) == 0 {
			return errors.New("volume name required")
		}

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
		"the required size of the volume in bytes")

	createVolumeCmd.Flags().Uint64Var(
		&createVolume.limBytes,
		"lim-bytes",
		0,
		"the limit to the size of the volume in bytes")

	createVolumeCmd.Flags().Var(
		&createVolume.caps,
		"cap",
		"one or more volume capabilities. "+
			"ex: --cap 1,block --cap 5,mount,xfs,uid=500")

	createVolumeCmd.Flags().Var(
		&createVolume.params,
		"params",
		"one or more volume parameter key/value pairs")

	createVolumeCmd.Flags().BoolVar(
		&root.withRequiresCreds,
		"with-requires-credentials",
		false,
		"marks the request's credentials as a required field")

	createVolumeCmd.Flags().BoolVar(
		&root.withRequiresVolumeAttributes,
		"with-requires-attributes",
		false,
		"marks the response's attributes as a required field")

	createVolumeCmd.Flags().BoolVar(
		&root.withSuccessCreateVolumeAlreadyExists,
		"with-success-already-exists",
		false,
		"treats an already exists error as success")
}
