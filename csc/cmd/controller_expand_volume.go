package cmd

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var expandVolume struct {
	reqBytes int64
	limBytes int64
	volCap   volumeCapabilitySliceArg
}

var expandVolumeCmd = &cobra.Command{
	Use:     "expand-volume",
	Aliases: []string{"exp", "expand"},
	Short:   `invokes the rpc "ControllerExpandVolume"`,
	Example: `
USAGE

    csc controller expand [flags] VOLUME_ID [VOLUME_ID...]
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		req := csi.ControllerExpandVolumeRequest{
			Secrets:          root.secrets,
			VolumeCapability: expandVolume.volCap.data[0],
		}

		if expandVolume.reqBytes > 0 || expandVolume.limBytes > 0 {
			req.CapacityRange = &csi.CapacityRange{}
			if v := expandVolume.reqBytes; v > 0 {
				req.CapacityRange.RequiredBytes = v
			}
			if v := expandVolume.limBytes; v > 0 {
				req.CapacityRange.LimitBytes = v
			}
		}

		for i := range args {
			ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
			defer cancel()

			// Set the volume name for the current request.
			req.VolumeId = args[i]

			log.WithField("request", req).Debug("expanding volume")
			rep, err := controller.client.ControllerExpandVolume(ctx, &req)
			if err != nil {
				return err
			}

			fmt.Println(rep.CapacityBytes)
		}

		return nil
	},
}

func init() {
	controllerCmd.AddCommand(expandVolumeCmd)

	flagRequiredBytes(expandVolumeCmd.Flags(), &expandVolume.reqBytes)

	flagLimitBytes(expandVolumeCmd.Flags(), &expandVolume.limBytes)

	flagVolumeCapability(expandVolumeCmd.Flags(), &expandVolume.volCap)

	flagWithRequiresCreds(
		expandVolumeCmd.Flags(),
		&root.withRequiresCreds,
		"")
}
