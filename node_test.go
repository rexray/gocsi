package gocsi_test

import (
	"context"
	"path"

	"google.golang.org/grpc"

	"github.com/thecodeteam/gocsi"
	"github.com/thecodeteam/gocsi/csi"
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
		version = mockSupportedVersions[0]
	})
	AfterEach(func() {
		ctx = nil
		gclient.Close()
		gclient = nil
		client = nil
		version = nil
		stopMock()
	})

	Describe("GetNodeID", func() {
		var nodeID string
		BeforeEach(func() {
			nodeID, err = gocsi.GetNodeID(
				ctx,
				client,
				mockSupportedVersions[0])
		})
		It("Should Be Valid", func() {
			Ω(err).ShouldNot(HaveOccurred())
			Ω(nodeID).ShouldNot(BeEmpty())
			Ω(nodeID).Should(Equal(service.Name))
		})
	})

	Describe("Publication", func() {

		device := "/dev/mock"
		targetPath := "/mnt/mock"
		mntPathKey := path.Join(service.Name, targetPath)

		publishVolume := func() {
			err = gocsi.NodePublishVolume(
				ctx,
				client,
				version,
				"1",
				nil,
				map[string]string{"device": device},
				targetPath,
				gocsi.NewMountCapability(
					csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					"mock",
					nil),
				false,
				nil)
		}

		shouldBePublished := func() {
			Ω(err).ShouldNot(HaveOccurred())
		}

		BeforeEach(func() {
			publishVolume()
		})
		Context("PublishVolume", func() {
			It("Should Be Valid", func() {
				shouldBePublished()
				vols, _, err := gocsi.ListVolumes(
					ctx, csi.NewControllerClient(gclient), version, 0, "")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(vols).Should(HaveLen(3))
				Ω(vols[0].Attributes[mntPathKey]).Should(Equal(device))
			})
		})

		Context("UnpublishVolume", func() {
			BeforeEach(func() {
				shouldBePublished()
				err := gocsi.NodeUnpublishVolume(
					ctx,
					client,
					version,
					"1",
					targetPath,
					nil)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("Should Be Unpublished", func() {
				vols, _, err := gocsi.ListVolumes(
					ctx, csi.NewControllerClient(gclient), version, 0, "")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(vols).Should(HaveLen(3))
				_, ok := vols[0].Attributes[mntPathKey]
				Ω(ok).Should(BeFalse())
			})
		})
	})
})
