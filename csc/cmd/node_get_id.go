package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var nodeGetIDCmd = &cobra.Command{
	Use:     "get-id",
	Aliases: []string{"id"},
	Short:   `invokes the rpc "NodeGetId"`,
	RunE: func(cmd *cobra.Command, args []string) error {

		ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
		defer cancel()

		rep, err := node.client.NodeGetId(
			ctx,
			&csi.NodeGetIdRequest{
				Version: &root.version.Version,
			})
		if err != nil {
			return err
		}

		fmt.Println(rep.NodeId)
		return nil
	},
}

func init() {
	nodeCmd.AddCommand(nodeGetIDCmd)

	nodeGetIDCmd.Flags().BoolVar(
		&root.withRequiresNodeID,
		"with-requires-node-id",
		false,
		"marks the response's node ID as a required field")
}
