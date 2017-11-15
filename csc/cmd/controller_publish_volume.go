package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/thecodeteam/gocsi"
	"github.com/thecodeteam/gocsi/csi"
)

var controllerPublishVolume struct {
	nodeID  string
	caps    volumeCapabilitySliceArg
	attribs mapOfStringArg
}

var controllerPublishVolumeCmd = &cobra.Command{
	Use:     "publishvolume",
	Aliases: []string{"a", "att", "pub", "attach", "publish"},
	Short:   `invokes the rpc "ControllerPublishVolume"`,
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) == 0 {
			return errors.New("volume ID required")
		}

		req := csi.ControllerPublishVolumeRequest{
			Version:          &root.version.Version,
			NodeId:           controllerPublishVolume.nodeID,
			UserCredentials:  gocsi.ParseMap(os.Getenv(userCredsKey)),
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
		"the id of the node to which to publish the volume")

	controllerPublishVolumeCmd.Flags().Var(
		&controllerPublishVolume.caps,
		"cap",
		"the volume capability to publish")

	controllerPublishVolumeCmd.Flags().Var(
		&controllerPublishVolume.attribs,
		"attrib",
		"one or more volume attributes key/value pairs")
}
