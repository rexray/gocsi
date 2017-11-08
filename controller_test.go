package gocsi_test

import (
	"context"
	"path"

	"google.golang.org/grpc"

	"github.com/thecodeteam/gocsi"
	"github.com/thecodeteam/gocsi/csi"
	"github.com/thecodeteam/gocsi/mock/service"
)

var _ = Describe("Controller", func() {
	var (
		err      error
		stopMock func()
		ctx      context.Context
		gclient  *grpc.ClientConn
		client   csi.ControllerClient

		version *csi.Version

		vol      *csi.VolumeInfo
		volName  string
		reqBytes uint64
		limBytes uint64
		fsType   string
		mntFlags []string
		params   map[string]string

		pubVolInfo map[string]string
	)
	BeforeEach(func() {
		ctx = context.Background()
		gclient, stopMock, err = startMockServer(ctx)
		Ω(err).ShouldNot(HaveOccurred())
		client = csi.NewControllerClient(gclient)

		version = mockSupportedVersions[0]

		volName = "Test Volume"
		reqBytes = 1.074e+10 //  10GiB
		limBytes = 1.074e+11 // 100GiB
		fsType = "ext4"
		mntFlags = []string{"-o noexec"}
		params = map[string]string{"tag": "gold"}
	})
	AfterEach(func() {
		ctx = nil
		gclient.Close()
		gclient = nil
		client = nil
		stopMock()

		version = nil

		vol = nil
		volName = ""
		reqBytes = 0
		limBytes = 0
		fsType = ""
		mntFlags = nil
		params = nil
		pubVolInfo = nil
	})

	createNewVolume := func() {
		vol, err = gocsi.CreateVolume(
			ctx,
			client,
			version,
			volName,
			reqBytes,
			limBytes,
			[]*csi.VolumeCapability{
				gocsi.NewMountCapability(0, fsType, mntFlags),
			},
			params)
	}

	validateNewVolume := func() {
		Ω(err).ShouldNot(HaveOccurred())
		Ω(vol).ShouldNot(BeNil())
		Ω(vol.CapacityBytes).Should(Equal(limBytes))
		Ω(vol.Attributes["name"]).Should(Equal(volName))
	}

	Describe("CreateVolume", func() {
		JustBeforeEach(func() {
			createNewVolume()
		})
		Context("Normal Create Volume Call", func() {
			It("Should Be Valid", validateNewVolume)
		})
		Context("No LimitBytes", func() {
			BeforeEach(func() {
				limBytes = 0
			})
			It("Should Be Valid", func() {
				Ω(err).ShouldNot(HaveOccurred())
				Ω(vol).ShouldNot(BeNil())
				Ω(vol.CapacityBytes).Should(Equal(reqBytes))
				Ω(vol.Attributes["name"]).Should(Equal(volName))
			})
		})
		Context("Missing Name", func() {
			BeforeEach(func() {
				volName = ""
			})
			It("Should Be Invalid", func() {
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Σ(&gocsi.Error{
					FullMethod:  "/csi.Controller/CreateVolume",
					Code:        3,
					Description: "name required",
				}))
				Ω(vol).Should(BeNil())
			})
		})
		Context("Idempotent Create", func() {
			It("Should Be Valid", func() {
				// Validate the new volume with this specific function.
				// It's the same function that will be used to validate
				// the volume that's created as the result of the
				// idempotent create.
				validateNewVolume()

				var (
					vols      []*csi.VolumeInfo
					nextToken string
				)

				// Verify that the newly created volume increases
				// the volume count to 4.
				listVolsAndValidate4 := func() {
					vols, nextToken, err = gocsi.ListVolumes(
						ctx,
						client,
						version,
						0,
						"")
					Ω(err).ShouldNot(HaveOccurred())
					Ω(vols).ShouldNot(BeNil())
					Ω(vols).Should(HaveLen(4))
				}
				listVolsAndValidate4()

				// Create the same volume again and then assert the
				// volume count has not increased.
				createNewVolume()
				validateNewVolume()
				listVolsAndValidate4()
			})
		})
	})

	Describe("DeleteVolume", func() {
		var volID string
		BeforeEach(func() {
			volID = CTest().ComponentTexts[2]
		})
		AfterEach(func() {
			volID = ""
		})
		JustBeforeEach(func() {
			err = gocsi.DeleteVolume(
				ctx,
				client,
				version,
				volID)
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
		Context("3", func() {
			It("Should Be Valid", func() {
				Ω(err).ShouldNot(HaveOccurred())
			})
		})
		Context("Missing Volume ID", func() {
			BeforeEach(func() {
				volID = ""
			})
			It("Should Not Be Valid", func() {
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Σ(&gocsi.Error{
					FullMethod:  "/csi.Controller/DeleteVolume",
					Code:        3,
					Description: "volume id required",
				}))
			})
		})
		Context("Missing Version", func() {
			BeforeEach(func() {
				version = nil
			})
			It("Should Not Be Valid", func() {
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Σ(&gocsi.Error{
					FullMethod:  "/csi.Controller/DeleteVolume",
					Code:        2,
					Description: "unsupported request version: 0.0.0",
				}))
			})
		})
	})

	Describe("ListVolumes", func() {
		var (
			vols          []*csi.VolumeInfo
			maxEntries    uint32
			startingToken string
			nextToken     string
		)
		AfterEach(func() {
			vols = nil
			maxEntries = 0
			startingToken = ""
			nextToken = ""
			version = nil
		})
		JustBeforeEach(func() {
			vols, nextToken, err = gocsi.ListVolumes(
				ctx,
				client,
				version,
				maxEntries,
				startingToken)
		})
		Context("Normal List Volumes Call", func() {
			It("Should Be Valid", func() {
				Ω(err).ShouldNot(HaveOccurred())
				Ω(vols).ShouldNot(BeNil())
				Ω(vols).Should(HaveLen(3))
			})
		})
		Context("Create Volume Then List", func() {
			BeforeEach(func() {
				createNewVolume()
				validateNewVolume()
			})
			It("Should Be Valid", func() {
				Ω(err).ShouldNot(HaveOccurred())
				Ω(vols).ShouldNot(BeNil())
				Ω(vols).Should(HaveLen(4))
			})
		})
	})

	Describe("Publication", func() {

		devPathKey := path.Join(service.Name, "dev")

		publishVolume := func() {
			pubVolInfo, err = gocsi.ControllerPublishVolume(
				ctx,
				client,
				version,
				"1",
				nil,
				service.Name,
				gocsi.NewMountCapability(
					csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					"mock",
					nil),
				true)
		}

		shouldBePublished := func() {
			Ω(err).ShouldNot(HaveOccurred())
			Ω(pubVolInfo).ShouldNot(BeNil())
			Ω(pubVolInfo["device"]).Should(Equal("/dev/mock"))
		}

		BeforeEach(func() {
			publishVolume()
		})
		Context("PublishVolume", func() {
			It("Should Be Valid", func() {
				shouldBePublished()
				vols, _, err := gocsi.ListVolumes(ctx, client, version, 0, "")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(vols).Should(HaveLen(3))
				Ω(vols[0].Attributes[devPathKey]).Should(Equal("/dev/mock"))
			})
		})

		Context("UnpublishVolume", func() {
			BeforeEach(func() {
				shouldBePublished()
				err := gocsi.ControllerUnpublishVolume(
					ctx,
					client,
					version,
					"1",
					service.Name)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("Should Be Unpublished", func() {
				vols, _, err := gocsi.ListVolumes(ctx, client, version, 0, "")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(vols).Should(HaveLen(3))
				_, ok := vols[0].Attributes[devPathKey]
				Ω(ok).Should(BeFalse())
			})
		})
	})
})
