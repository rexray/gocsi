package cmd

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var controllerPublishVolume struct {
	nodeID  string
	caps    volumeCapabilitySliceArg
	attribs mapOfStringArg
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
			Version:          &root.version.Version,
			NodeId:           controllerPublishVolume.nodeID,
			UserCredentials:  root.userCreds,
			VolumeAttributes: controllerPublishVolume.attribs.data,
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
			for k, v := range rep.PublishVolumeInfo {
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

	controllerPublishVolumeCmd.Flags().BoolVar(
		&root.withRequiresPubVolInfo,
		"with-requires-pub-info",
		false,
		`Marks the response's PublishVolumeInfo field as required.
        Enabling this option also enables --with-spec-validation.`)

	flagVolumeAttributes(
		controllerPublishVolumeCmd.Flags(), &controllerPublishVolume.attribs)

	flagVolumeCapability(
		controllerPublishVolumeCmd.Flags(), &controllerPublishVolume.caps)

	flagWithRequiresCreds(
		controllerPublishVolumeCmd.Flags(),
		&root.withRequiresCreds,
		"")

	flagWithRequiresAttribs(
		controllerPublishVolumeCmd.Flags(),
		&root.withRequiresVolumeAttributes,
		"")
}
