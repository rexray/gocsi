package cmd

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var nodeStageVolume struct {
	nodeID            string
	stagingTargetPath string
	pubCtx            mapOfStringArg
	volCtx            mapOfStringArg
	caps              volumeCapabilitySliceArg
}

var nodeStageVolumeCmd = &cobra.Command{
	Use:   "stage",
	Short: `invokes the rpc "NodeStageVolume"`,
	Example: `
USAGE

    csc node stage [flags] VOLUME_ID [VOLUME_ID...]
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		req := csi.NodeStageVolumeRequest{
			StagingTargetPath: nodeStageVolume.stagingTargetPath,
			PublishContext:    nodeStageVolume.pubCtx.data,
			Secrets:           root.secrets,
			VolumeContext:     nodeStageVolume.volCtx.data,
		}

		if len(nodeStageVolume.caps.data) > 0 {
			req.VolumeCapability = nodeStageVolume.caps.data[0]
		}

		for i := range args {
			ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
			defer cancel()

			// Set the volume ID for the current request.
			req.VolumeId = args[i]

			log.WithField("request", req).Debug("staging volume")
			_, err := node.client.NodeStageVolume(ctx, &req)
			if err != nil {
				return err
			}

			fmt.Println(args[i])
		}

		return nil
	},
}

func init() {
	nodeCmd.AddCommand(nodeStageVolumeCmd)

	flagStagingTargetPath(
		nodeStageVolumeCmd.Flags(), &nodeStageVolume.stagingTargetPath)

	flagVolumeCapability(
		nodeStageVolumeCmd.Flags(), &nodeStageVolume.caps)

	flagVolumeContext(nodeStageVolumeCmd.Flags(), &nodeStageVolume.volCtx)

	flagPublishContext(nodeStageVolumeCmd.Flags(), &nodeStageVolume.pubCtx)

	flagWithRequiresCreds(
		nodeStageVolumeCmd.Flags(), &root.withRequiresCreds, "")

	flagWithRequiresVolContext(
		nodeStageVolumeCmd.Flags(), &root.withRequiresVolContext, false)

	flagWithRequiresPubContext(
		nodeStageVolumeCmd.Flags(), &root.withRequiresPubContext, false)
}
