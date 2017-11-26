package cmd

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var valVolCaps struct {
	attribs mapOfStringArg
	caps    volumeCapabilitySliceArg
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
			Version:            &root.version.Version,
			VolumeAttributes:   valVolCaps.attribs.data,
			VolumeCapabilities: valVolCaps.caps.data,
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
			fmt.Printf("%q\t%v", args[i], rep.Supported)
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

	flagVolumeAttributes(valVolCapsCmd.Flags(), &valVolCaps.attribs)

	flagVolumeCapabilities(valVolCapsCmd.Flags(), &valVolCaps.caps)

	flagWithRequiresAttribs(
		valVolCapsCmd.Flags(),
		&root.withRequiresVolumeAttributes,
		"")
}
