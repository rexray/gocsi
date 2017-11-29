package gocsi_test

import (
	"context"
	"math"
	"path"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/thecodeteam/gocsi"
	"github.com/thecodeteam/gocsi/csp"
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

		vol        *csi.VolumeInfo
		volID      string
		volName    string
		reqBytes   uint64
		limBytes   uint64
		fsType     string
		mntFlags   []string
		params     map[string]string
		userCreds  map[string]string
		pubVolInfo map[string]string
	)
	BeforeEach(func() {
		ctx = context.Background()
		version = &mockSupportedVersions[0]
		volID = "4"
		volName = "Test Volume"
		reqBytes = 1.074e+10 //  10GiB
		limBytes = 1.074e+11 // 100GiB
		fsType = "ext4"
		mntFlags = []string{"-o noexec"}
		params = map[string]string{"tag": "gold"}
		userCreds = map[string]string{"beour": "guest"}
	})
	JustBeforeEach(func() {
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

	listVolumes := func() (vols []csi.VolumeInfo, err error) {
		cvol, cerr := gocsi.PageVolumes(
			ctx,
			client,
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

	createNewVolumeWithResult := func() (*csi.VolumeInfo, error) {
		req := &csi.CreateVolumeRequest{
			Name:    volName,
			Version: version,
			CapacityRange: &csi.CapacityRange{
				RequiredBytes: reqBytes,
				LimitBytes:    limBytes,
			},
			VolumeCapabilities: []*csi.VolumeCapability{
				gocsi.NewMountCapability(0, fsType, mntFlags...),
			},
			UserCredentials: userCreds,
			Parameters:      params,
		}
		res, err := client.CreateVolume(ctx, req)
		if res == nil {
			return nil, err
		}
		return res.VolumeInfo, err
	}

	createNewVolume := func() {
		vol, err = createNewVolumeWithResult()
	}

	validateNewVolumeResult := func(
		vol *csi.VolumeInfo,
		err error) bool {

		if err != nil {
			Ω(err).Should(Σ(gocsi.ErrOpPending))
			return true
		}

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
				Ω(err).Should(Σ(gocsi.ErrVolumeNameRequired))
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
				vols, err := listVolumes()
				Ω(err).ShouldNot(HaveOccurred())
				Ω(vols).Should(HaveLen(4))
			}

			idempCreateVols := func() {
				var (
					once    sync.Once
					buckets = count / bucketSize
					worker  = func() {
						defer wg.Done()
						defer GinkgoRecover()
						if !validateNewVolumeResult(
							createNewVolumeWithResult()) {
							once.Do(func() {
								opPendingErrorOccurs = true
							})
						}
					}
				)
				if r := math.Remainder(
					float64(count), float64(bucketSize)); r > 0 {
					buckets++
				}
				//fmt.Fprintf(
				//	GinkgoWriter, "count=%d, buckets=%d\n", count, buckets)
				for i := 0; i < buckets; i++ {
					go func(i int) {
						defer GinkgoRecover()
						start := i * bucketSize
						for j := start; j < start+bucketSize && j < count; j++ {
							//fmt.Fprintf(
							//	GinkgoWriter, "bucket=%d, index=%d\n", i, j)
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
			Context("x10", func() {
				BeforeEach(func() {
					count = 10
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
		deleteVolume := func() {
			_, err = client.DeleteVolume(
				ctx,
				&csi.DeleteVolumeRequest{
					Version:  version,
					VolumeId: volID,
				})
		}
		JustBeforeEach(func() {
			deleteVolume()
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
				Ω(err).Should(Σ(gocsi.ErrVolumeIDRequired))
			})
		})
		Context("Not Found", func() {
			BeforeEach(func() {
				volID = "5"
			})

			It("Should Be Valid", func() {
				Ω(err).ShouldNot(HaveOccurred())
				deleteVolume()
				Ω(err).ShouldNot(HaveOccurred())
				deleteVolume()
				Ω(err).ShouldNot(HaveOccurred())
				deleteVolume()
				Ω(err).ShouldNot(HaveOccurred())
			})
			Context("With NotFound Error", func() {
				BeforeEach(func() {
					ctx = gocsi.WithEnviron(ctx,
						[]string{
							csp.EnvVarDeleteVolNotFoundSuccess + "=false",
						})
				})
				shouldNotBeFound := func() {
					Ω(err).Should(HaveOccurred())
					stat, ok := status.FromError(err)
					Ω(ok).Should(BeTrue())
					Ω(stat).ShouldNot(BeNil())
					Ω(stat.Code()).Should(Equal(codes.NotFound))
				}
				It("Should Be Valid", func() {
					shouldNotBeFound()
					deleteVolume()
					shouldNotBeFound()
					deleteVolume()
					shouldNotBeFound()
					deleteVolume()
					shouldNotBeFound()
				})
			})
		})
		Context("Missing Version", func() {
			BeforeEach(func() {
				version = nil
			})
			It("Should Not Be Valid", func() {
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(ΣCM(
					codes.InvalidArgument,
					"invalid request version: nil"))
			})
		})
	})

	Describe("ListVolumes", func() {
		var vols []csi.VolumeInfo
		AfterEach(func() {
			vols = nil
		})
		JustBeforeEach(func() {
			vols, err = listVolumes()
		})
		Context("Normal List Volumes Call", func() {
			It("Should Be Valid", func() {
				Ω(err).ShouldNot(HaveOccurred())
				Ω(vols).ShouldNot(BeNil())
				Ω(vols).Should(HaveLen(3))
			})
		})
		Context("Create Volume Then List", func() {
			JustBeforeEach(func() {
				createNewVolume()
				validateNewVolume()
				vols, err = listVolumes()
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
			req := &csi.ControllerPublishVolumeRequest{
				Version:  version,
				VolumeId: "1",
				NodeId:   service.Name,
				Readonly: true,
				VolumeCapability: gocsi.NewMountCapability(
					csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					"mock"),
			}
			res, err := client.ControllerPublishVolume(ctx, req)
			Ω(err).ShouldNot(HaveOccurred())
			pubVolInfo = res.PublishVolumeInfo
		}

		shouldBePublished := func() {
			Ω(err).ShouldNot(HaveOccurred())
			Ω(pubVolInfo).ShouldNot(BeNil())
			Ω(pubVolInfo["device"]).Should(Equal("/dev/mock"))
		}

		JustBeforeEach(func() {
			publishVolume()
		})
		Context("PublishVolume", func() {
			It("Should Be Valid", func() {
				shouldBePublished()
				vols, err := listVolumes()
				Ω(err).ShouldNot(HaveOccurred())
				Ω(vols).Should(HaveLen(3))
				Ω(vols[0].Attributes[devPathKey]).Should(Equal("/dev/mock"))
			})
		})

		Context("UnpublishVolume", func() {
			JustBeforeEach(func() {
				shouldBePublished()
				_, err := client.ControllerUnpublishVolume(
					ctx,
					&csi.ControllerUnpublishVolumeRequest{
						Version:  version,
						VolumeId: "1",
						NodeId:   service.Name,
					})
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("Should Be Unpublished", func() {
				vols, err := listVolumes()
				Ω(err).ShouldNot(HaveOccurred())
				Ω(vols).Should(HaveLen(3))
				_, ok := vols[0].Attributes[devPathKey]
				Ω(ok).Should(BeFalse())
			})
		})
	})
})
