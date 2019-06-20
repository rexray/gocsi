package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var getCapacity struct {
	caps   volumeCapabilitySliceArg
	params mapOfStringArg
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
				VolumeCapabilities: getCapacity.caps.data,
				Parameters:         getCapacity.params.data,
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
	flagParameters(getCapacityCmd.Flags(), &getCapacity.params)
}
