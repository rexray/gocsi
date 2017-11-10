package gocsi_test

import (
	"context"
	"fmt"
	"math"
	"path"
	"sync"

	"google.golang.org/grpc"

	"github.com/thecodeteam/gocsi"
	"github.com/thecodeteam/gocsi/csi"
	"github.com/thecodeteam/gocsi/mock/service"
)

var _ = Describe("Controller", func() {
	var (
		err      error
		gocsiErr error
		stopMock func()
		ctx      context.Context
		gclient  *grpc.ClientConn
		client   csi.ControllerClient

		version *csi.Version

		vol       *csi.VolumeInfo
		volID     string
		volName   string
		reqBytes  uint64
		limBytes  uint64
		fsType    string
		mntFlags  []string
		params    map[string]string
		userCreds map[string]string

		pubVolInfo map[string]string
	)
	BeforeEach(func() {
		ctx = context.Background()
		gclient, stopMock, err = startMockServer(ctx)
		Ω(err).ShouldNot(HaveOccurred())
		client = csi.NewControllerClient(gclient)

		gocsiErr = &gocsi.Error{}
		version = mockSupportedVersions[0]

		volID = "4"
		volName = "Test Volume"
		reqBytes = 1.074e+10 //  10GiB
		limBytes = 1.074e+11 // 100GiB
		fsType = "ext4"
		mntFlags = []string{"-o noexec"}
		params = map[string]string{"tag": "gold"}
		userCreds = map[string]string{"beour": "guest"}
	})
	AfterEach(func() {
		ctx = nil
		gclient.Close()
		gclient = nil
		client = nil
		stopMock()

		gocsiErr = nil
		version = nil

		vol = nil
		volID = ""
		volName = ""
		reqBytes = 0
		limBytes = 0
		fsType = ""
		mntFlags = nil
		params = nil
		pubVolInfo = nil
	})

	createNewVolumeWithResult := func() (*csi.VolumeInfo, error) {
		return gocsi.CreateVolume(
			ctx,
			client,
			version,
			volName,
			reqBytes,
			limBytes,
			[]*csi.VolumeCapability{
				gocsi.NewMountCapability(0, fsType, mntFlags),
			},
			userCreds,
			params)
	}

	createNewVolume := func() {
		vol, err = createNewVolumeWithResult()
	}

	validateNewVolumeResult := func(
		vol *csi.VolumeInfo,
		err error) bool {

		if err != nil {
			Ω(vol).Should(BeNil())
			Ω(err).Should(BeAssignableToTypeOf(gocsiErr))
			terr := err.(*gocsi.Error)
			Ω(terr.Code).Should(BeEquivalentTo(
				csi.Error_CreateVolumeError_OPERATION_PENDING_FOR_VOLUME))
			return true
		}

		Ω(err).ShouldNot(HaveOccurred())
		Ω(vol).ShouldNot(BeNil())
		Ω(vol.CapacityBytes).Should(Equal(limBytes))
		Ω(vol.Id).Should(Equal(volID))
		Ω(vol.Attributes["name"]).Should(Equal(volName))
		return false
	}

	validateNewVolume := func() {
		validateNewVolumeResult(vol, err)
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
					FullMethod:  gocsi.FMCreateVolume,
					Code:        3,
					Description: "name required",
				}))
				Ω(vol).Should(BeNil())
			})
		})
		Context("Idempotent Create", func() {

			const bucketSize = 250

			var (
				wg                   sync.WaitGroup
				count                int
				opPendingErrorOccurs bool
			)

			// Verify that the newly created volume increases
			// the volume count to 4.
			listVolsAndValidate4 := func() {
				vols, _, err := gocsi.ListVolumes(
					ctx,
					client,
					version,
					0,
					"")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(vols).ShouldNot(BeNil())
				Ω(vols).Should(HaveLen(4))
			}

			idempCreateVols := func() {
				var (
					once    sync.Once
					buckets = count / bucketSize
					worker  = func() {
						defer GinkgoRecover()
						if !validateNewVolumeResult(
							createNewVolumeWithResult()) {
							once.Do(func() {
								opPendingErrorOccurs = true
							})
						}
						wg.Done()
					}
				)
				if r := math.Remainder(
					float64(count), float64(bucketSize)); r > 0 {
					buckets++
				}
				fmt.Fprintf(
					GinkgoWriter, "count=%d, buckets=%d\n", count, buckets)
				for i := 0; i < buckets; i++ {
					go func(i int) {
						defer GinkgoRecover()
						start := i * bucketSize
						for j := start; j < start+bucketSize && j < count; j++ {
							fmt.Fprintf(
								GinkgoWriter, "bucket=%d, index=%d\n", i, j)
							go worker()
						}
					}(i)
				}
			}

			validateIdempResult := func() {
				wg.Wait()
				if count >= 1000 {
					Ω(opPendingErrorOccurs).Should(BeTrue())
				}
				listVolsAndValidate4()
			}

			JustBeforeEach(func() {
				validateNewVolume()
				listVolsAndValidate4()
				idempCreateVols()
				wg.Add(count)
			})

			AfterEach(func() {
				count = 0
				opPendingErrorOccurs = false
			})

			Context("x1", func() {
				BeforeEach(func() {
					count = 1
				})
				It("Should Be Valid", validateIdempResult)
			})
			Context("x1000", func() {
				BeforeEach(func() {
					count = 1000
				})
				It("Should Be Valid", validateIdempResult)
			})
			Context("x10000", func() {
				BeforeEach(func() {
					count = 10000
				})
				It("Should Be Valid", validateIdempResult)
			})
			Context("x100000", func() {
				BeforeEach(func() {
					count = 100000
				})
				It("Should Be Valid", validateIdempResult)
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
				volID,
				nil)
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
					FullMethod:  gocsi.FMDeleteVolume,
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
					FullMethod:  gocsi.FMDeleteVolume,
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
				true,
				nil)
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
					service.Name,
					nil)
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

type createVolumeResult struct {
	vol *csi.VolumeInfo
	err error
}
