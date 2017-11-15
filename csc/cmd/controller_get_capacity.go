package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/thecodeteam/gocsi/csi"
)

var getCapacity struct {
	caps volumeCapabilitySliceArg
}

var getCapacityCmd = &cobra.Command{
	Use:     "getcapacity",
	Aliases: []string{"getcapac", "capac"},
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

	getCapacityCmd.Flags().Var(
		&getCapacity.caps,
		"cap",
		"one or more volume capabilities. "+
			"ex: --cap 1,block --cap 5,mount,xfs,uid=500")
}
