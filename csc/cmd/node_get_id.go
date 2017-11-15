package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/thecodeteam/gocsi/csi"
)

var nodeGetIDCmd = &cobra.Command{
	Use:     "getid",
	Aliases: []string{"id"},
	Short:   `invokes the rpc "GetNodeID"`,
	RunE: func(cmd *cobra.Command, args []string) error {

		ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
		defer cancel()

		rep, err := node.client.GetNodeID(
			ctx,
			&csi.GetNodeIDRequest{
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
}
