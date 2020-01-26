package cmd

import (
	"context"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var nodeGetVolumeStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: `invokes the rpc "NodeGetVolumeStats"`,
	Example: `
USAGE

	csc node stats VOLUME_ID:VOLUME_PATH:STAGING_PATH [VOLUME_ID:VOLUME_PATH:STAGING_PATH...]
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		req := csi.NodeGetVolumeStatsRequest{}

		for i := range args {
			ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
			defer cancel()

			// Set the volume ID and volume path for the current request.
			split := strings.Split(args[i], ":")
			req.VolumeId, req.VolumePath = split[0], split[1]
			if len(split) > 2 {
				req.StagingTargetPath = split[2]
			}

			log.WithField("request", req).Debug("staging volume")
			rep, err := node.client.NodeGetVolumeStats(ctx, &req)
			if err != nil {
				return err
			}
			if err := root.tpl.Execute(os.Stdout, struct {
				Name string
				Path string
				Resp *csi.NodeGetVolumeStatsResponse
			}{
				Name: req.VolumeId,
				Path: req.VolumePath,
				Resp: rep,
			}); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	nodeCmd.AddCommand(nodeGetVolumeStatsCmd)
}
