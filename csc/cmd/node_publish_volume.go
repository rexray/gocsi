package cmd

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var nodePublishVolume struct {
	nodeID     string
	targetPath string
	pubInfo    mapOfStringArg
	attribs    mapOfStringArg
	readOnly   bool
	caps       volumeCapabilitySliceArg
}

var nodePublishVolumeCmd = &cobra.Command{
	Use:     "publish",
	Aliases: []string{"mnt", "mount"},
	Short:   `invokes the rpc "NodePublishVolume"`,
	Example: `
USAGE

    csc node publishvolume [flags] VOLUME_ID [VOLUME_ID...]
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		req := csi.NodePublishVolumeRequest{
			Version:                &root.version.Version,
			TargetPath:             nodePublishVolume.targetPath,
			PublishInfo:            nodePublishVolume.pubInfo.data,
			Readonly:               nodePublishVolume.readOnly,
			NodePublishCredentials: root.userCreds,
			VolumeAttributes:       nodePublishVolume.attribs.data,
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
		&nodePublishVolume.targetPath,
		"target-path",
		"",
		"The path to which to mount the volume")

	nodePublishVolumeCmd.Flags().Var(
		&nodePublishVolume.pubInfo,
		"pub-info",
		`One or more key/value pairs may be specified to send with
        the request as its PublishVolumeInfo field:

                --pub-info key1=val1,key2=val2 --pub-infoparams=key3=val3`)

	nodePublishVolumeCmd.Flags().BoolVar(
		&nodePublishVolume.readOnly,
		"read-only",
		false,
		"Mark the volume as read-only")

	nodePublishVolumeCmd.Flags().BoolVar(
		&root.withRequiresPubVolInfo,
		"with-requires-pub-info",
		false,
		`Marks the request's PublishVolumeInfo field as required.
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
