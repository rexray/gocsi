package cmd

import (
	"context"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var createSnapshot struct {
	sourceVol string
	params    mapOfStringArg
}

var createSnapshotCmd = &cobra.Command{
	Use:     "create-snapshot",
	Aliases: []string{"s", "snap"},
	Short:   `invokes the rpc "CreateSnapshot"`,
	Example: `
CREATING MULTIPLE SNAPSHOTS
        The following example illustrates how to create two snapshots with the
        same characteristics at the same time:

            csc controller snap --endpoint /csi/server.sock
							    --source-vol MySourceVolume
                                MyNewSnapshot1 MyNewSnapshot2
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		req := csi.CreateSnapshotRequest{
			SourceVolumeId: createSnapshot.sourceVol,
			Parameters:     createSnapshot.params.data,
			Secrets:        root.secrets,
		}

		for i := range args {
			ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
			defer cancel()

			// Set the volume name for the current request.
			req.Name = args[i]
			if createSnapshot.sourceVol == "" {
				return fmt.Errorf("--source-volume MUST be provided")
			}

			log.WithField("request", req).Debug("creating snapshot")
			rep, err := controller.client.CreateSnapshot(ctx, &req)
			if err != nil {
				return err
			}
			if err := root.tpl.Execute(os.Stdout, rep.Snapshot); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	controllerCmd.AddCommand(createSnapshotCmd)

	createSnapshotCmd.Flags().StringVar(
		&createSnapshot.sourceVol,
		"source-volume",
		"",
		"The source volume to snapshot")

	flagParameters(createSnapshotCmd.Flags(), &createSnapshot.params)

	flagWithRequiresCreds(
		createSnapshotCmd.Flags(),
		&root.withRequiresCreds,
		"")
}
