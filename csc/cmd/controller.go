package cmd

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/spf13/cobra"
)

var controller struct {
	client csi.ControllerClient
}

// controllerCmd represents the controller command
var controllerCmd = &cobra.Command{
	Use:     "controller",
	Aliases: []string{"c"},
	Short:   "the csi controller service rpcs",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if f := cmd.Root().PersistentPreRunE; f != nil {
			if err := f(cmd, args); err != nil {
				return err
			}
		}
		controller.client = csi.NewControllerClient(root.client)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(controllerCmd)
}
