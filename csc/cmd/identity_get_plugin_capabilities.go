package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
)

var pluginCapsCmd = &cobra.Command{
	Use:     "plugin-capabilities",
	Aliases: []string{"caps"},
	Short:   `invokes the rpc "GetPluginCapabilities"`,
	RunE: func(cmd *cobra.Command, args []string) error {

		ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
		defer cancel()

		rep, err := identity.client.GetPluginCapabilities(
			ctx,
			&csi.GetPluginCapabilitiesRequest{})
		if err != nil {
			return err
		}

		return root.tpl.Execute(os.Stdout, rep)
	},
}

func init() {
	identityCmd.AddCommand(pluginCapsCmd)
}
