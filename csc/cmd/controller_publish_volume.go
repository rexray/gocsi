package cmd

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var controllerPublishVolume struct {
	nodeID   string
	caps     volumeCapabilitySliceArg
	volCtx   mapOfStringArg
	readOnly bool
}

var controllerPublishVolumeCmd = &cobra.Command{
	Use:     "publish",
	Aliases: []string{"attach"},
	Short:   `invokes the rpc "ControllerPublishVolume"`,
	Example: `
USAGE

    csc controller publishvolume [flags] VOLUME_ID [VOLUME_ID...]
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		req := csi.ControllerPublishVolumeRequest{
			NodeId:        controllerPublishVolume.nodeID,
			Secrets:       root.secrets,
			VolumeContext: controllerPublishVolume.volCtx.data,
			Readonly:      controllerPublishVolume.readOnly,
		}

		if len(controllerPublishVolume.caps.data) > 0 {
			req.VolumeCapability = controllerPublishVolume.caps.data[0]
		}

		for i := range args {
			ctx, cancel := context.WithTimeout(root.ctx, root.timeout)
			defer cancel()

			// Set the volume ID for the current request.
			req.VolumeId = args[i]

			log.WithField("request", req).Debug("publishing volume")
			rep, err := controller.client.ControllerPublishVolume(ctx, &req)
			if err != nil {
				return err
			}

			fmt.Printf("%q", args[i])
			for k, v := range rep.PublishContext {
				fmt.Printf("\t%q=%q", k, v)
			}
			fmt.Println()
		}

		return nil
	},
}

func init() {
	controllerCmd.AddCommand(controllerPublishVolumeCmd)

	controllerPublishVolumeCmd.Flags().StringVar(
		&controllerPublishVolume.nodeID,
		"node-id",
		"",
		"The ID of the node to which to publish the volume")

	controllerPublishVolumeCmd.Flags().BoolVar(
		&root.withRequiresNodeID,
		"with-requires-node-id",
		false,
		`Marks the request's NodeId field as required.
        Enabling this option also enables --with-spec-validation.`)

	flagReadOnly(
		controllerPublishVolumeCmd.Flags(), &controllerPublishVolume.readOnly)

	flagVolumeContext(
		controllerPublishVolumeCmd.Flags(), &controllerPublishVolume.volCtx)

	flagVolumeCapability(
		controllerPublishVolumeCmd.Flags(), &controllerPublishVolume.caps)

	flagWithRequiresCreds(
		controllerPublishVolumeCmd.Flags(),
		&root.withRequiresCreds,
		"")

	flagWithRequiresVolContext(
		controllerPublishVolumeCmd.Flags(), &root.withRequiresVolContext, false)

	flagWithRequiresPubContext(
		controllerPublishVolumeCmd.Flags(), &root.withRequiresPubContext, false)
}
