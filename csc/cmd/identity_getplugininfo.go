package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/thecodeteam/gocsi/csi"
)

var pluginInfoCmd = &cobra.Command{
	Use:     "plugininfo",
	Aliases: []string{"info", "getp"},
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
