package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var nodeProbeCmd = &cobra.Command{
	Use:   "probe",
	Short: `invokes the rpc "NodeProbe"`,
	RunE: func(cmd *cobra.Command, args []string) error {

		ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
		defer cancel()

		if _, err := node.client.NodeProbe(
			ctx,
			&csi.NodeProbeRequest{
				Version: &root.version.Version,
			}); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	nodeCmd.AddCommand(nodeProbeCmd)
}
