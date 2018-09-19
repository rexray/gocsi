package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"

	"github.com/rexray/gocsi/utils"
)

var listSnapshots struct {
	maxEntries    int32
	startingToken string
	paging        bool
}

var listSnapshotsCmd = &cobra.Command{
	Use:     "list-snapshots",
	Aliases: []string{"sl", "snap-list", "snapshots"},
	Short:   `invokes the rpc "ListSnapshots"`,
	RunE: func(*cobra.Command, []string) error {
		ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
		defer cancel()

		req := csi.ListSnapshotsRequest{
			MaxEntries:    listSnapshots.maxEntries,
			StartingToken: listSnapshots.startingToken,
		}

		// If auto-paging is not enabled then send a normal request.
		if !listSnapshots.paging {
			rep, err := controller.client.ListSnapshots(ctx, &req)
			if err != nil {
				return err
			}
			return root.tpl.Execute(os.Stdout, rep)
		}

		// Paging is enabled.
		cvol, cerr := utils.PageSnapshots(ctx, controller.client, req)
		for {
			select {
			case v, ok := <-cvol:
				if !ok {
					return nil
				}
				if err := root.tpl.Execute(os.Stdout, v); err != nil {
					return err
				}
			case e, ok := <-cerr:
				if !ok {
					return nil
				}
				return e
			}
		}
	},
}

func init() {
	controllerCmd.AddCommand(listSnapshotsCmd)

	listSnapshotsCmd.Flags().Int32Var(
		&listSnapshots.maxEntries,
		"max-entries",
		0,
		"The maximum number of entries to return")

	listSnapshotsCmd.Flags().StringVar(
		&listSnapshots.startingToken,
		"starting-token",
		"",
		"The starting token used to retrieve paged data")

	listSnapshotsCmd.Flags().BoolVar(
		&listSnapshots.paging,
		"paging",
		false,
		"Enables auto-paging")

	listSnapshotsCmd.Flags().StringVar(
		&root.format,
		"format",
		"",
		"The Go template format used to emit the results")
}
