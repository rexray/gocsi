package cmd

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/thecodeteam/gocsi/csi"
)

var deleteVolumeCmd = &cobra.Command{
	Use:     "deletevolume",
	Aliases: []string{"d", "del", "rm", "delete"},
	Short:   `invokes the rpc "DeleteVolume"`,
	Example: `
USAGE

    csc controller deletevolume [flags] VOLUME_ID [VOLUME_ID...]
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		req := csi.DeleteVolumeRequest{
			Version:         &root.version.Version,
			UserCredentials: root.userCreds,
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

	deleteVolumeCmd.Flags().BoolVar(
		&root.withRequiresCreds,
		"with-requires-credentials",
		false,
		"marks the request's credentials as a required field")

	deleteVolumeCmd.Flags().BoolVar(
		&root.withSuccessDeleteVolumeNotFound,
		"with-success-already-exists",
		false,
		"treats a not found error as success")
}
