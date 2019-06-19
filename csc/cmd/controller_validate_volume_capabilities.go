package cmd

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var valVolCaps struct {
	volCtx mapOfStringArg
	params mapOfStringArg
	caps   volumeCapabilitySliceArg
}

var valVolCapsCmd = &cobra.Command{
	Use:     "validate-volume-capabilities",
	Aliases: []string{"validate"},
	Short:   `invokes the rpc "ValidateVolumeCapabilities"`,
	Example: `
USAGE

    csc controller validate [flags] VOLUME_ID [VOLUME_ID...]
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		req := csi.ValidateVolumeCapabilitiesRequest{
			VolumeContext:      valVolCaps.volCtx.data,
			VolumeCapabilities: valVolCaps.caps.data,
			Parameters:         valVolCaps.params.data,
		}

		for i := range args {
			ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
			defer cancel()

			// Set the volume name for the current request.
			req.VolumeId = args[i]

			log.WithField("request", req).Debug("validate volume capabilities")
			rep, err := controller.client.ValidateVolumeCapabilities(ctx, &req)
			if err != nil {
				return err
			}
			fmt.Printf("%q\t%v", args[i], rep.Confirmed)
			if rep.Message != "" {
				fmt.Printf("\t%q", rep.Message)
			}
			fmt.Println()
		}

		return nil
	},
}

func init() {
	controllerCmd.AddCommand(valVolCapsCmd)

	flagParameters(valVolCapsCmd.Flags(), &valVolCaps.params)

	flagVolumeCapabilities(valVolCapsCmd.Flags(), &valVolCaps.caps)

	flagVolumeContext(valVolCapsCmd.Flags(), &valVolCaps.volCtx)

	flagWithRequiresVolContext(
		valVolCapsCmd.Flags(), &root.withRequiresVolContext, false)
}
