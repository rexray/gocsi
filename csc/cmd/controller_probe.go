package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var controllerProbeCmd = &cobra.Command{
	Use:   "probe",
	Short: `invokes the rpc "ControllerProbe"`,
	RunE: func(cmd *cobra.Command, args []string) error {

		ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
		defer cancel()

		if _, err := controller.client.ControllerProbe(
			ctx,
			&csi.ControllerProbeRequest{
				Version: &root.version.Version,
			}); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	controllerCmd.AddCommand(controllerProbeCmd)
}
