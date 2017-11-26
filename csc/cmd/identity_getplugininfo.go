package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var pluginInfoCmd = &cobra.Command{
	Use:     "plugin-info",
	Aliases: []string{"info"},
	Short:   `invokes the rpc "GetPluginInfo"`,
	RunE: func(cmd *cobra.Command, args []string) error {

		ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
		defer cancel()

		rep, err := identity.client.GetPluginInfo(
			ctx,
			&csi.GetPluginInfoRequest{
				Version: &root.version.Version,
			})
		if err != nil {
			return err
		}

		return root.tpl.Execute(os.Stdout, rep)
	},
}

func init() {
	identityCmd.AddCommand(pluginInfoCmd)
}
