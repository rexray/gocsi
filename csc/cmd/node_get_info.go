package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
)

var nodeGetInfoCmd = &cobra.Command{
	Use:     "get-info",
	Aliases: []string{"info"},
	Short:   `invokes the rpc "NodeGetInfo"`,
	RunE: func(cmd *cobra.Command, args []string) error {

		ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
		defer cancel()

		rep, err := node.client.NodeGetInfo(ctx, &csi.NodeGetInfoRequest{})
		if err != nil {
			return err
		}

		return root.tpl.Execute(os.Stdout, rep)
	},
}

func init() {
	nodeCmd.AddCommand(nodeGetInfoCmd)

	nodeGetInfoCmd.Flags().BoolVar(
		&root.withRequiresNodeID,
		"with-requires-node-id",
		false,
		"marks the response's node ID as a required field")
}
