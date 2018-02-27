package cmd

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
)

var nodeUnstageVolume struct {
	stagingTargetPath string
}

var nodeUnstageVolumeCmd = &cobra.Command{
	Use:   "unstage",
	Short: `invokes the rpc "NodeUnstageVolume"`,
	Example: `
USAGE

    csc node unstage [flags] VOLUME_ID [VOLUME_ID...]
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		req := csi.NodeUnstageVolumeRequest{
			StagingTargetPath: nodeUnstageVolume.stagingTargetPath,
		}

		for i := range args {
			ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
			defer cancel()

			// Set the volume ID for the current request.
			req.VolumeId = args[i]

			log.WithField("request", req).Debug("unstaging volume")
			_, err := node.client.NodeUnstageVolume(ctx, &req)
			if err != nil {
				return err
			}

			fmt.Println(args[i])
		}

		return nil
	},
}

func init() {
	nodeCmd.AddCommand(nodeUnstageVolumeCmd)

	nodeUnstageVolumeCmd.Flags().StringVar(
		&nodeUnstageVolume.stagingTargetPath,
		"staging-target-path",
		"",
		"The path from which to unstage the volume")
}
