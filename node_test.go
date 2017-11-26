package gocsi_test

import (
	"context"
	"path"

	"google.golang.org/grpc"

	"github.com/thecodeteam/gocsi"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/thecodeteam/gocsi/mock/service"
)

var _ = Describe("Node", func() {
	var (
		err      error
		stopMock func()
		ctx      context.Context
		gclient  *grpc.ClientConn
		client   csi.NodeClient
		version  *csi.Version
	)
	BeforeEach(func() {
		ctx = context.Background()
		gclient, stopMock, err = startMockServer(ctx)
		Ω(err).ShouldNot(HaveOccurred())
		client = csi.NewNodeClient(gclient)
		version = &mockSupportedVersions[0]
	})
	AfterEach(func() {
		ctx = nil
		gclient.Close()
		gclient = nil
		client = nil
		version = nil
		stopMock()
	})

	listVolumes := func() (vols []csi.VolumeInfo, err error) {
		cvol, cerr := gocsi.PageVolumes(
			ctx,
			csi.NewControllerClient(gclient),
			csi.ListVolumesRequest{Version: version})
		for {
			select {
			case v, ok := <-cvol:
				if !ok {
					return
				}
				vols = append(vols, v)
			case e, ok := <-cerr:
				if !ok {
					return
				}
				err = e
			}
		}
	}

	Describe("GetNodeID", func() {
		var nodeID string
		BeforeEach(func() {
			res, err := client.GetNodeID(
				ctx,
				&csi.GetNodeIDRequest{
					Version: &mockSupportedVersions[0],
				})
			Ω(err).ShouldNot(HaveOccurred())
			nodeID = res.NodeId
		})
		It("Should Be Valid", func() {
			Ω(nodeID).ShouldNot(BeEmpty())
			Ω(nodeID).Should(Equal(service.Name))
		})
	})

	Describe("Publication", func() {

		device := "/dev/mock"
		targetPath := "/mnt/mock"
		mntPathKey := path.Join(service.Name, targetPath)

		publishVolume := func() {
			req := &csi.NodePublishVolumeRequest{
				Version:           version,
				VolumeId:          "1",
				PublishVolumeInfo: map[string]string{"device": device},
				VolumeCapability: gocsi.NewMountCapability(
					csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					"mock"),
				TargetPath: targetPath,
			}
			_, err = client.NodePublishVolume(ctx, req)
			Ω(err).ShouldNot(HaveOccurred())
		}

		BeforeEach(func() {
			publishVolume()
		})
		Context("PublishVolume", func() {
			It("Should Be Valid", func() {
				vols, err := listVolumes()
				Ω(err).ShouldNot(HaveOccurred())
				Ω(vols).Should(HaveLen(3))
				Ω(vols[0].Attributes[mntPathKey]).Should(Equal(device))
			})
		})

		Context("UnpublishVolume", func() {
			BeforeEach(func() {
				_, err = client.NodeUnpublishVolume(
					ctx,
					&csi.NodeUnpublishVolumeRequest{
						Version:    version,
						VolumeId:   "1",
						TargetPath: targetPath,
					})
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("Should Be Unpublished", func() {
				vols, err := listVolumes()
				Ω(err).ShouldNot(HaveOccurred())
				Ω(vols).Should(HaveLen(3))
				_, ok := vols[0].Attributes[mntPathKey]
				Ω(ok).Should(BeFalse())
			})
		})
	})
})
