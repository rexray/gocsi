package cmd

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
)

var nodeStageVolume struct {
	nodeID            string
	stagingTargetPath string
	pubInfo           mapOfStringArg
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
			PublishInfo:       nodeStageVolume.pubInfo.data,
			NodeStageSecrets:  root.secrets,
			VolumeAttributes:  nodeStageVolume.attribs.data,
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
		&nodeStageVolume.pubInfo,
		"pub-info",
		`One or more key/value pairs may be specified to send with
        the request as its PublishInfo field:

                --pub-info key1=val1,key2=val2 --pub-infoparams=key3=val3`)

	nodeStageVolumeCmd.Flags().BoolVar(
		&root.withRequiresPubVolInfo,
		"with-requires-pub-info",
		false,
		`Marks the request's PublishInfo field as required.
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
