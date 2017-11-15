package cmd

import (
	"context"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/thecodeteam/gocsi/csi"
)

var deleteVolumeCmd = &cobra.Command{
	Use:     "deletevolume",
	Aliases: []string{"d", "del", "rm", "delete"},
	Short:   `invokes the rpc "DeleteVolume"`,
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) == 0 {
			return errors.New("volume ID required")
		}

		req := csi.DeleteVolumeRequest{
			Version: &root.version.Version,
		}

		for i := range args {
			ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
			defer cancel()

			// Set the volume ID for the current request.
			req.VolumeId = args[i]

			log.WithField("request", req).Debug("deleting volume")
			_, err := controller.client.DeleteVolume(ctx, &req)
			if err != nil {
				return err
			}
			fmt.Println(args[i])
		}

		return nil
	},
}

func init() {
	controllerCmd.AddCommand(deleteVolumeCmd)
}
