package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
)

var probeCmd = &cobra.Command{
	Use:   "probe",
	Short: `invokes the rpc "Probe"`,
	RunE: func(cmd *cobra.Command, args []string) error {

		ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
		defer cancel()

		rep, err := identity.client.Probe(
			ctx,
			&csi.ProbeRequest{})
		if err != nil {
			return err
		}

		return root.tpl.Execute(os.Stdout, rep)
	},
}

func init() {
	identityCmd.AddCommand(probeCmd)
}
