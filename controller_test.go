package gocsi_test

import (
	"context"
	"errors"

	"google.golang.org/grpc"

	"github.com/codedellemc/gocsi"
	"github.com/codedellemc/gocsi/csi"
)

var _ = Describe("Controller", func() {
	var (
		err      error
		stopMock func()
		ctx      context.Context
		gclient  *grpc.ClientConn
		client   csi.ControllerClient
	)
	BeforeEach(func() {
		ctx = context.Background()
		gclient, stopMock, err = startMockServer(ctx)
		Ω(err).ShouldNot(HaveOccurred())
		client = csi.NewControllerClient(gclient)
	})
	AfterEach(func() {
		ctx = nil
		gclient.Close()
		gclient = nil
		client = nil
		stopMock()
	})

	Describe("DeleteVolume", func() {
		var (
			volID   *csi.VolumeID
			version *csi.Version
		)
		BeforeEach(func() {
			version = mockSupportedVersions[0]
			volID = &csi.VolumeID{
				Values: map[string]string{
					"id": CTest().ComponentTexts[2],
				},
			}
		})
		AfterEach(func() {
			volID = nil
			version = nil
		})
		JustBeforeEach(func() {
			err = gocsi.DeleteVolume(
				ctx,
				client,
				version,
				volID,
				&csi.VolumeMetadata{Values: map[string]string{}})
		})
		Context("0", func() {
			It("Should Be Valid", func() {
				Ω(err).ShouldNot(HaveOccurred())
			})
		})
		Context("1", func() {
			It("Should Be Valid", func() {
				Ω(err).ShouldNot(HaveOccurred())
			})
		})
		Context("2", func() {
			It("Should Be Valid", func() {
				Ω(err).ShouldNot(HaveOccurred())
			})
		})
		Context("Missing Volume ID", func() {
			BeforeEach(func() {
				volID = nil
			})
			It("Should Not Be Valid", func() {
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(errors.New(
					"error: DeleteVolume failed: 3: missing id obj")))
			})
		})
		Context("Missing Version", func() {
			BeforeEach(func() {
				version = nil
			})
			It("Should Not Be Valid", func() {
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(errors.New(
					"error: DeleteVolume failed: 2: " +
						"unsupported request version: 0.0.0")))
			})
		})
	})
})
