package cmd

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var controllerUnpublishVolume struct {
	nodeID string
}

var controllerUnpublishVolumeCmd = &cobra.Command{
	Use:     "unpublish",
	Aliases: []string{"detach"},
	Short:   `invokes the rpc "ControllerUnpublishVolume"`,
	Example: `
USAGE

    csc controller unpublishvolume [flags] VOLUME_ID [VOLUME_ID...]
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		req := csi.ControllerUnpublishVolumeRequest{
			Version: &root.version.Version,
			NodeId:  controllerUnpublishVolume.nodeID,
			ControllerUnpublishCredentials: root.userCreds,
		}

		for i := range args {
			ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
			defer cancel()

			// Set the volume ID for the current request.
			req.VolumeId = args[i]

			log.WithField("request", req).Debug("unpublishing volume")
			_, err := controller.client.ControllerUnpublishVolume(ctx, &req)
			if err != nil {
				return err
			}
			if err != nil {
				return err
			}
			fmt.Println(args[i])
		}

		return nil
	},
}

func init() {
	controllerCmd.AddCommand(controllerUnpublishVolumeCmd)

	controllerUnpublishVolumeCmd.Flags().StringVar(
		&controllerUnpublishVolume.nodeID,
		"node-id",
		"",
		"The ID of the node from which to unpublish the volume")

	flagWithRequiresCreds(
		controllerUnpublishVolumeCmd.Flags(),
		&root.withRequiresCreds,
		"")
}
