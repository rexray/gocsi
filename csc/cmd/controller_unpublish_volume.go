package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/thecodeteam/gocsi"
	"github.com/thecodeteam/gocsi/csi"
)

var controllerUnpublishVolume struct {
	nodeID string
}

var controllerUnpublishVolumeCmd = &cobra.Command{
	Use: "unpublishvolume",
	Aliases: []string{
		"u", "det", "dett", "upub", "unpub", "detach", "unpublish"},
	Short: `invokes the rpc "ControllerUnpublishVolume"`,
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) == 0 {
			return errors.New("volume ID required")
		}

		req := csi.ControllerUnpublishVolumeRequest{
			Version:         &root.version.Version,
			NodeId:          controllerUnpublishVolume.nodeID,
			UserCredentials: gocsi.ParseMap(os.Getenv(userCredsKey)),
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
		"the id of the node to which to publish the volume")
}
