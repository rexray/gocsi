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
	attribs           mapOfStringArg
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
			VolumeContext:     nodeStageVolume.attribs.data,
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

	nodeStageVolumeCmd.Flags().StringVar(
		&nodeStageVolume.stagingTargetPath,
		"staging-target-path",
		"",
		"The path to which to stage the volume")

	nodeStageVolumeCmd.Flags().Var(
		&nodeStageVolume.pubCtx,
		"pub-context",
		`One or more key/value pairs may be specified to send with
        the request as its PublishContext field:

                --pub-context key1=val1,key2=val2`)

	nodeStageVolumeCmd.Flags().BoolVar(
		&root.withRequiresPubVolContext,
		"with-requires-pub-context",
		false,
		`Marks the request's PublisContext field as required.
        Enabling this option also enables --with-spec-validation.`)

	flagVolumeAttributes(
		nodeStageVolumeCmd.Flags(), &nodeStageVolume.attribs)

	flagVolumeCapability(
		nodeStageVolumeCmd.Flags(), &nodeStageVolume.caps)

	flagWithRequiresCreds(
		nodeStageVolumeCmd.Flags(), &root.withRequiresCreds, "")

	flagWithRequiresAttribs(
		nodeStageVolumeCmd.Flags(), &root.withRequiresVolumeAttributes, "")
}
