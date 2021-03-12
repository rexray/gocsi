package gocsi_test

import (
	"context"
	"path"

	"google.golang.org/grpc"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/dell/gocsi/mock/service"
	"github.com/dell/gocsi/utils"
)

const (
	kib    int64 = 1024
	mib    int64 = kib * 1024
	gib    int64 = mib * 1024
	gib100 int64 = gib * 100
	tib    int64 = gib * 1024
	tib100 int64 = tib * 100
)

var _ = Describe("Node", func() {
	var (
		err      error
		stopMock func()
		ctx      context.Context
		gclient  *grpc.ClientConn
		client   csi.NodeClient
	)
	BeforeEach(func() {
		ctx = context.Background()
		gclient, stopMock, err = startMockServer(ctx)
		Ω(err).ShouldNot(HaveOccurred())
		client = csi.NewNodeClient(gclient)
	})
	AfterEach(func() {
		ctx = nil
		gclient.Close()
		gclient = nil
		client = nil
		stopMock()
	})

	listVolumes := func() (vols []csi.Volume, err error) {
		cvol, cerr := utils.PageVolumes(
			ctx,
			csi.NewControllerClient(gclient),
			csi.ListVolumesRequest{})
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

	Describe("NodeGetVolumeStats", func() {
		var (
			volID   = "Mock Volume 2"
			volPath = "/root/mock-vol"
		)
		BeforeEach(func() {
			resp, err := client.NodeGetVolumeStats(
				ctx,
				&csi.NodeGetVolumeStatsRequest{
					VolumeId:   volID,
					VolumePath: volPath,
				})
			usage := resp.Usage[0]
			Ω(err).ShouldNot(HaveOccurred())
			Ω(usage.Total).Should(Equal(gib100))
			Ω(usage.Used).Should(Equal(int64(float64(gib100) * 0.4)))
			Ω(usage.Available).Should(Equal(int64(float64(gib100) * 0.6)))
		})
	})

	Describe("NodeGetInfo", func() {
		var nodeID string
		var maxVolsPerNode int64
		BeforeEach(func() {
			res, err := client.NodeGetInfo(
				ctx,
				&csi.NodeGetInfoRequest{})
			Ω(err).ShouldNot(HaveOccurred())
			nodeID = res.GetNodeId()
			maxVolsPerNode = res.GetMaxVolumesPerNode()
		})
		It("Should Be Valid", func() {
			Ω(nodeID).ShouldNot(BeEmpty())
			Ω(nodeID).Should(Equal(service.Name))
			Ω(maxVolsPerNode).Should(Equal(int64(0)))
		})
	})

	Describe("Publication", func() {

		device := "/dev/mock"
		targetPath := "/mnt/mock"
		mntPathKey := path.Join(service.Name, targetPath)

		publishVolume := func() {
			req := &csi.NodePublishVolumeRequest{
				VolumeId:       "1",
				PublishContext: map[string]string{"device": device},
				VolumeCapability: utils.NewMountCapability(
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
				Ω(vols[0].VolumeContext[mntPathKey]).Should(Equal(device))
			})
		})

		Context("UnpublishVolume", func() {
			BeforeEach(func() {
				_, err = client.NodeUnpublishVolume(
					ctx,
					&csi.NodeUnpublishVolumeRequest{
						VolumeId:   "1",
						TargetPath: targetPath,
					})
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("Should Be Unpublished", func() {
				vols, err := listVolumes()
				Ω(err).ShouldNot(HaveOccurred())
				Ω(vols).Should(HaveLen(3))
				_, ok := vols[0].VolumeContext[mntPathKey]
				Ω(ok).Should(BeFalse())
			})
		})
	})
})
