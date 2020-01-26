package cmd

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var nodeExpandVolume struct {
	reqBytes    int64
	limBytes    int64
	stagingPath string
	volCap      *volumeCapabilitySliceArg
}

var nodeExpandVolumeCmd = &cobra.Command{
	Use:     "expand-volume",
	Aliases: []string{"exp", "expand"},
	Short:   `invokes the rpc "NodeExpandVolume"`,
	Example: `
USAGE

    csc node expand [flags] VOLUME_ID VOLUME_PATH
`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {

		// Set the volume name and path for the current request.
		req := csi.NodeExpandVolumeRequest{
			VolumeId:          args[0],
			VolumePath:        args[1],
			StagingTargetPath: nodeExpandVolume.stagingPath,
			VolumeCapability:  nodeExpandVolume.volCap.data[0],
		}

		if nodeExpandVolume.reqBytes > 0 || nodeExpandVolume.limBytes > 0 {
			req.CapacityRange = &csi.CapacityRange{}
			if v := nodeExpandVolume.reqBytes; v > 0 {
				req.CapacityRange.RequiredBytes = v
			}
			if v := nodeExpandVolume.limBytes; v > 0 {
				req.CapacityRange.LimitBytes = v
			}
		}

		ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
		defer cancel()

		log.WithField("request", req).Debug("expanding volume")
		rep, err := node.client.NodeExpandVolume(ctx, &req)
		if err != nil {
			return err
		}

		fmt.Println(rep.CapacityBytes)

		return nil
	},
}

func init() {
	nodeCmd.AddCommand(nodeExpandVolumeCmd)

	flagRequiredBytes(nodeExpandVolumeCmd.Flags(), &nodeExpandVolume.reqBytes)

	flagLimitBytes(nodeExpandVolumeCmd.Flags(), &nodeExpandVolume.limBytes)

	flagStagingTargetPath(nodeExpandVolumeCmd.Flags(), &nodeExpandVolume.stagingPath)

	flagVolumeCapability(nodeExpandVolumeCmd.Flags(), nodeExpandVolume.volCap)
}
