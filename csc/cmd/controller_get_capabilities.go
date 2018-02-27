package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
)

var controllerGetCapabilitiesCmd = &cobra.Command{
	Use:     "get-capabilities",
	Aliases: []string{"capabilities"},
	Short:   `invokes the rpc "ControllerGetCapabilities"`,
	RunE: func(cmd *cobra.Command, args []string) error {

		ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
		defer cancel()

		rep, err := controller.client.ControllerGetCapabilities(
			ctx,
			&csi.ControllerGetCapabilitiesRequest{})
		if err != nil {
			return err
		}

		for _, cap := range rep.Capabilities {
			fmt.Println(cap.Type)
		}

		return nil
	},
}

func init() {
	controllerCmd.AddCommand(controllerGetCapabilitiesCmd)
}
