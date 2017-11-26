//go:generate make

// Package gocsi provides a Container Storage Interface (CSI) library,
// client, and other helpful utilities.
package gocsi

import (
	"sync"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

const (
	// Namespace is the namesapce used by the protobuf.
	Namespace = "csi"

	// CSIEndpoint is the name of the environment variable that
	// contains the CSI endpoint.
	CSIEndpoint = "CSI_ENDPOINT"

	//
	// Controller Service
	//
	ctrlSvc = "/" + Namespace + ".Controller/"

	// CreateVolume is the full method name for the
	// eponymous RPC message.
	CreateVolume = ctrlSvc + "CreateVolume"

	// DeleteVolume is the full method name for the
	// eponymous RPC message.
	DeleteVolume = ctrlSvc + "DeleteVolume"

	// ControllerPublishVolume is the full method name for the
	// eponymous RPC message.
	ControllerPublishVolume = ctrlSvc + "ControllerPublishVolume"

	// ControllerUnpublishVolume is the full method name for the
	// eponymous RPC message.
	ControllerUnpublishVolume = ctrlSvc + "ControllerUnpublishVolume"

	// ValidateVolumeCapabilities is the full method name for the
	// eponymous RPC message.
	ValidateVolumeCapabilities = ctrlSvc + "ValidateVolumeCapabilities"

	// ListVolumes is the full method name for the
	// eponymous RPC message.
	ListVolumes = ctrlSvc + "ListVolumes"

	// GetCapacity is the full method name for the
	// eponymous RPC message.
	GetCapacity = ctrlSvc + "GetCapacity"

	// ControllerGetCapabilities is the full method name for the
	// eponymous RPC message.
	ControllerGetCapabilities = ctrlSvc + "ControllerGetCapabilities"

	// ControllerProbe is the full method name for the
	// eponymous RPC message.
	ControllerProbe = ctrlSvc + "ControllerProbe"

	//
	// Identity Service
	//
	identSvc = "/" + Namespace + ".Identity/"

	// GetSupportedVersions is the full method name for the
	// eponymous RPC message.
	GetSupportedVersions = identSvc + "GetSupportedVersions"

	// GetPluginInfo is the full method name for the
	// eponymous RPC message.
	GetPluginInfo = identSvc + "GetPluginInfo"

	//
	// Node Service
	//
	nodeSvc = "/" + Namespace + ".Node/"

	// GetNodeID is the full method name for the
	// eponymous RPC message.
	GetNodeID = nodeSvc + "GetNodeID"

	// NodePublishVolume is the full method name for the
	// eponymous RPC message.
	NodePublishVolume = nodeSvc + "NodePublishVolume"

	// NodeUnpublishVolume is the full method name for the
	// eponymous RPC message.
	NodeUnpublishVolume = nodeSvc + "NodeUnpublishVolume"

	// NodeProbe is the full method name for the
	// eponymous RPC message.
	NodeProbe = nodeSvc + "NodeProbe"

	// NodeGetCapabilities is the full method name for the
	// eponymous RPC message.
	NodeGetCapabilities = nodeSvc + "NodeGetCapabilities"
)

// NewMountCapability returns a new *csi.VolumeCapability for a
// volume that is to be mounted.
func NewMountCapability(
	mode csi.VolumeCapability_AccessMode_Mode,
	fsType string,
	mountFlags ...string) *csi.VolumeCapability {

	return &csi.VolumeCapability{
		AccessMode: &csi.VolumeCapability_AccessMode{
			Mode: mode,
		},
		AccessType: &csi.VolumeCapability_Mount{
			Mount: &csi.VolumeCapability_MountVolume{
				FsType:     fsType,
				MountFlags: mountFlags,
			},
		},
	}
}

// NewBlockCapability returns a new *csi.VolumeCapability for a
// volume that is to be accessed as a raw device.
func NewBlockCapability(
	mode csi.VolumeCapability_AccessMode_Mode) *csi.VolumeCapability {

	return &csi.VolumeCapability{
		AccessMode: &csi.VolumeCapability_AccessMode{
			Mode: mode,
		},
		AccessType: &csi.VolumeCapability_Block{
			Block: &csi.VolumeCapability_BlockVolume{},
		},
	}
}

// PageVolumes issues one or more ListVolumes requests to retrieve
// all available volumes, returning them over a Go channel.
func PageVolumes(
	ctx context.Context,
	client csi.ControllerClient,
	req csi.ListVolumesRequest,
	opts ...grpc.CallOption) (<-chan csi.VolumeInfo, <-chan error) {

	var (
		cvol = make(chan csi.VolumeInfo)
		cerr = make(chan error)
	)

	// Execute the RPC in a goroutine, looping until there are no
	// more volumes available.
	go func() {
		var (
			wg     sync.WaitGroup
			pages  int
			cancel context.CancelFunc
		)

		// Get a cancellation context used to control the interaction
		// between returning volumes and the possibility of an error.
		ctx, cancel = context.WithCancel(ctx)

		// waitAndClose closes the volume and error channels after all
		// channel-dependent goroutines have completed their work
		defer func() {
			wg.Wait()
			close(cerr)
			close(cvol)
			log.WithField("pages", pages).Debug("PageAllVolumes: exit")
		}()

		sendVolumes := func(res csi.ListVolumesResponse) {
			// Loop over the volume entries until they're all gone
			// or the context is cancelled.
			var i int
			for i = 0; i < len(res.Entries) && ctx.Err() == nil; i++ {

				// Send the volume over the channel.
				cvol <- *res.Entries[i].VolumeInfo

				// Let the wait group know that this worker has completed
				// its task.
				wg.Done()
			}
			// If not all volumes have been sent over the channel then
			// deduct the remaining number from the wait group.
			if i != len(res.Entries) {
				rem := len(res.Entries) - i
				log.WithFields(map[string]interface{}{
					"cancel":    ctx.Err(),
					"remaining": rem,
				}).Warn("PageAllVolumes: cancelled w unprocessed results")
				wg.Add(-rem)
			}
		}

		// listVolumes returns true if there are more volumes to list.
		listVolumes := func() bool {

			// The wait group "wg" is blocked during the execution of
			// this function.
			wg.Add(1)
			defer wg.Done()

			res, err := client.ListVolumes(ctx, &req, opts...)
			if err != nil {
				cerr <- err

				// Invoke the cancellation context function to
				// ensure that work wraps up as quickly as possible.
				cancel()

				return false
			}

			// Add to the number of workers
			wg.Add(len(res.Entries))

			// Process the retrieved volumes.
			go sendVolumes(*res)

			// Set the request's starting token to the response's
			// next token.
			req.StartingToken = res.NextToken
			return req.StartingToken != ""
		}

		// List volumes until there are no more volumes or the context
		// is cancelled.
		for {
			if ctx.Err() != nil {
				break
			}
			if !listVolumes() {
				break
			}
			pages++
		}
	}()

	return cvol, cerr
}
