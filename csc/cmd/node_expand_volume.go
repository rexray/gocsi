package cmd

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var nodeExpandVolume struct {
	reqBytes int64
	limBytes int64
	volPath  string
}

var nodeExpandVolumeCmd = &cobra.Command{
	Use:     "expand-volume",
	Aliases: []string{"exp", "expand"},
	Short:   `invokes the rpc "NodeExpandVolume"`,
	Example: `
USAGE

    csc node expand [flags] VOLUME_ID
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		req := csi.NodeExpandVolumeRequest{
			VolumePath: nodeExpandVolume.volPath,
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

		// Set the volume name for the current request.
		req.VolumeId = args[0]

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
	controllerCmd.AddCommand(expandVolumeCmd)

	flagRequiredBytes(nodeExpandVolumeCmd.Flags(), &nodeExpandVolume.reqBytes)

	flagLimitBytes(nodeExpandVolumeCmd.Flags(), &nodeExpandVolume.limBytes)

}
