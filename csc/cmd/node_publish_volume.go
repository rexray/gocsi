package cmd

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/thecodeteam/gocsi/csi"
)

var nodePublishVolume struct {
	nodeID     string
	targetPath string
	pubVolInfo mapOfStringArg
	attribs    mapOfStringArg
	readOnly   bool
	caps       volumeCapabilitySliceArg
}

var nodePublishVolumeCmd = &cobra.Command{
	Use:     "publishvolume",
	Aliases: []string{"pub", "mnt", "mount", "publish"},
	Short:   `invokes the rpc "NodePublishVolume"`,
	Example: `
USAGE

    csc node publishvolume [flags] VOLUME_ID [VOLUME_ID...]
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		req := csi.NodePublishVolumeRequest{
			Version:           &root.version.Version,
			TargetPath:        nodePublishVolume.targetPath,
			PublishVolumeInfo: nodePublishVolume.pubVolInfo.data,
			Readonly:          nodePublishVolume.readOnly,
			UserCredentials:   root.userCreds,
			VolumeAttributes:  nodePublishVolume.attribs.data,
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
		"the path to which to mount the volume")

	nodePublishVolumeCmd.Flags().Var(
		&nodePublishVolume.pubVolInfo,
		"pub-info",
		"one or more publication info key/value pairs")

	nodePublishVolumeCmd.Flags().Var(
		&nodePublishVolume.attribs,
		"attrib",
		"one or more volume attributes key/value pairs")

	nodePublishVolumeCmd.Flags().Var(
		&nodePublishVolume.caps,
		"cap",
		"the volume capability to publish")

	nodePublishVolumeCmd.Flags().BoolVar(
		&nodePublishVolume.readOnly,
		"read-only",
		false,
		"a flag that indicates whether or not the volume is read-only")

	nodePublishVolumeCmd.Flags().BoolVar(
		&root.withRequiresCreds,
		"with-requires-credentials",
		false,
		"marks the request's credentials as a required field")

	nodePublishVolumeCmd.Flags().BoolVar(
		&root.withRequiresPubVolInfo,
		"with-requires-pub-info",
		false,
		"marks the request's publish volume info as a required field")

	nodePublishVolumeCmd.Flags().BoolVar(
		&root.withRequiresVolumeAttributes,
		"with-requires-attributes",
		false,
		"marks the request's attributes as a required field")
}
