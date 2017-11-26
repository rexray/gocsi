package cmd

import (
	"github.com/spf13/cobra"
	"github.com/container-storage-interface/spec/lib/go/csi"
)

var node struct {
	client csi.NodeClient
}

// nodeCmd represents the node command
var nodeCmd = &cobra.Command{
	Use:     "node",
	Aliases: []string{"n"},
	Short:   "the csi node service rpcs",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if f := cmd.Root().PersistentPreRunE; f != nil {
			if err := f(cmd, args); err != nil {
				return err
			}
		}
		node.client = csi.NewNodeClient(root.client)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(nodeCmd)
}
