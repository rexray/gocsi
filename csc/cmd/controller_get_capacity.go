package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var getCapacity struct {
	caps volumeCapabilitySliceArg
}

var getCapacityCmd = &cobra.Command{
	Use:     "get-capacity",
	Aliases: []string{"capacity"},
	Short:   `invokes the rpc "GetCapacity"`,
	RunE: func(cmd *cobra.Command, args []string) error {

		ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
		defer cancel()

		rep, err := controller.client.GetCapacity(
			ctx,
			&csi.GetCapacityRequest{
				Version:            &root.version.Version,
				VolumeCapabilities: getCapacity.caps.data,
			})
		if err != nil {
			return err
		}

		fmt.Println(rep.AvailableCapacity)
		return nil
	},
}

func init() {
	controllerCmd.AddCommand(getCapacityCmd)
	flagVolumeCapabilities(getCapacityCmd.Flags(), &getCapacity.caps)
}
