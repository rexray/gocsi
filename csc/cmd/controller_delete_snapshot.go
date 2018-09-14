package cmd

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
)

var deleteSnapshotCmd = &cobra.Command{
	Use:     "delete-snapshot",
	Aliases: []string{"ds", "delsnap"},
	Short:   `invokes the rpc "DeleteSnapshot"`,
	Example: `
USAGE

    csc controller delete-snapshot [flags] snapshot_ID [snapshot_ID...]
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		req := csi.DeleteSnapshotRequest{
			DeleteSnapshotSecrets: root.secrets,
		}

		for i := range args {
			ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
			defer cancel()

			// Set the snapshot ID for the current request.
			req.SnapshotId = args[i]

			log.WithField("request", req).Debug("deleting snapshot")
			_, err := controller.client.DeleteSnapshot(ctx, &req)
			if err != nil {
				return err
			}
			fmt.Println(args[i])
		}

		return nil
	},
}

func init() {
	controllerCmd.AddCommand(deleteSnapshotCmd)

	flagWithRequiresCreds(
		deleteSnapshotCmd.Flags(),
		&root.withRequiresCreds,
		"")
}
