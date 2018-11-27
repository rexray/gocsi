package cmd

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/spf13/cobra"
)

var identity struct {
	client csi.IdentityClient
}

// identityCmd represents the controller command
var identityCmd = &cobra.Command{
	Use:     "identity",
	Aliases: []string{"i", "ident"},
	Short:   "the csi identity service rpcs",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if f := cmd.Root().PersistentPreRunE; f != nil {
			if err := f(cmd, args); err != nil {
				return err
			}
		}
		identity.client = csi.NewIdentityClient(root.client)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(identityCmd)
}
