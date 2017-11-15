package cmd

import (
	"context"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/thecodeteam/gocsi/csi"
)

var valVolCaps struct {
	attribs mapOfStringArg
	caps    volumeCapabilitySliceArg
}

var valVolCapsCmd = &cobra.Command{
	Use:     "validatevolumecapabilities",
	Aliases: []string{"v", "vv", "vvc", "validate"},
	Short:   `invokes the rpc "ValidateVolumeCapabilities"`,
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) == 0 {
			return errors.New("volume ID required")
		}

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

	valVolCapsCmd.Flags().Var(
		&valVolCaps.caps,
		"cap",
		"one or more volume capabilities. "+
			"ex: --cap 1,block --cap 5,mount,xfs,uid=500")

	valVolCapsCmd.Flags().Var(
		&valVolCaps.attribs,
		"attrib",
		"one or more volume attributes key/value pairs")
}
