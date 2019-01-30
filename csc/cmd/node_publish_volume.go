package cmd

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var nodePublishVolume struct {
	targetPath        string
	stagingTargetPath string
	pubCtx            mapOfStringArg
	attribs           mapOfStringArg
	readOnly          bool
	caps              volumeCapabilitySliceArg
}

var nodePublishVolumeCmd = &cobra.Command{
	Use:     "publish",
	Aliases: []string{"mnt", "mount"},
	Short:   `invokes the rpc "NodePublishVolume"`,
	Example: `
USAGE

    csc node publish [flags] VOLUME_ID [VOLUME_ID...]
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		req := csi.NodePublishVolumeRequest{
			StagingTargetPath: nodePublishVolume.stagingTargetPath,
			TargetPath:        nodePublishVolume.targetPath,
			PublishContext:    nodePublishVolume.pubCtx.data,
			Readonly:          nodePublishVolume.readOnly,
			Secrets:           root.secrets,
			VolumeContext:     nodePublishVolume.attribs.data,
		}

		if len(nodePublishVolume.caps.data) > 0 {
			req.VolumeCapability = nodePublishVolume.caps.data[0]
		}

		for i := range args {
			ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
			defer cancel()

			// Set the volume ID for the current request.
			req.VolumeId = args[i]

			log.WithField("request", req).Debug("mounting volume")
			_, err := node.client.NodePublishVolume(ctx, &req)
			if err != nil {
				return err
			}

			fmt.Println(args[i])
		}

		return nil
	},
}

func init() {
	nodeCmd.AddCommand(nodePublishVolumeCmd)

	nodePublishVolumeCmd.Flags().StringVar(
		&nodePublishVolume.stagingTargetPath,
		"staging-target-path",
		"",
		"The path from which to bind mount the volume")

	nodePublishVolumeCmd.Flags().StringVar(
		&nodePublishVolume.targetPath,
		"target-path",
		"",
		"The path to which to mount the volume")

	nodePublishVolumeCmd.Flags().Var(
		&nodePublishVolume.pubCtx,
		"pub-context",
		`One or more key/value pairs may be specified to send with
        the request as its PublishContext field:

                --pub-context key1=val1,key2=val2`)

	nodePublishVolumeCmd.Flags().BoolVar(
		&nodePublishVolume.readOnly,
		"read-only",
		false,
		"Mark the volume as read-only")

	nodePublishVolumeCmd.Flags().BoolVar(
		&root.withRequiresPubVolContext,
		"with-requires-pub-info",
		false,
		`Marks the request's PublishContext field as required.
        Enabling this option also enables --with-spec-validation.`)

	flagVolumeAttributes(
		nodePublishVolumeCmd.Flags(), &nodePublishVolume.attribs)

	flagVolumeCapability(
		nodePublishVolumeCmd.Flags(), &nodePublishVolume.caps)

	flagWithRequiresCreds(
		nodePublishVolumeCmd.Flags(), &root.withRequiresCreds, "")

	flagWithRequiresAttribs(
		nodePublishVolumeCmd.Flags(), &root.withRequiresVolumeAttributes, "")
}
