package gocsi_test

import (
	"context"
	"fmt"
	"math"
	"path"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/container-storage-interface/spec/lib/go/csi"

	"github.com/dell/gocsi/mock/service"
	"github.com/dell/gocsi/utils"
)

var _ = Describe("Controller", func() {
	var (
		err      error
		stopMock func()
		ctx      context.Context
		gclient  *grpc.ClientConn
		client   csi.ControllerClient

		vol      *csi.Volume
		snap     *csi.Snapshot
		volID    string
		snapID   string
		volName  string
		snapName string
		reqBytes int64
		limBytes int64
		expBytes int64
		capBytes int64
		fsType   string
		mntFlags []string
		params   map[string]string
		// userCreds map[string]string
		pubInfo map[string]string
		nodeId  string
	)
	BeforeEach(func() {
		ctx = context.Background()
		volID = "4"
		snapID = "12"
		volName = "Test Volume"
		snapName = "Test Snap"
		reqBytes = 1.074e+10 //  10GiB
		limBytes = 1.074e+11 // 100GiB
		expBytes = 1.074e+12 //   1TiB
		fsType = "ext4"
		mntFlags = []string{"-o noexec"}
		params = map[string]string{"tag": "gold"}
		// userCreds = map[string]string{"beour": "guest"}
		nodeId = service.Name
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

		vol = nil
		volID = ""
		volName = ""
		snapName = ""
		reqBytes = 0
		limBytes = 0
		expBytes = 0
		capBytes = 0
		fsType = ""
		mntFlags = nil
		params = nil
		pubInfo = nil
		nodeId = ""
	})

	listVolumes := func() (vols []csi.Volume, err error) {
		cvol, cerr := utils.PageVolumes(
			ctx,
			client,
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

	listSnapshots := func() (snaps []csi.Snapshot, err error) {
		csnap, cerr := utils.PageSnapshots(
			ctx,
			client,
			csi.ListSnapshotsRequest{})
		for {
			select {
			case s, ok := <-csnap:
				if !ok {
					return
				}
				snaps = append(snaps, s)
			case e, ok := <-cerr:
				if !ok {
					return
				}
				err = e
			}
		}
	}

	createNewVolumeWithResult := func() (*csi.Volume, error) {
		req := &csi.CreateVolumeRequest{
			Name: volName,
			CapacityRange: &csi.CapacityRange{
				RequiredBytes: reqBytes,
				LimitBytes:    limBytes,
			},
			VolumeCapabilities: []*csi.VolumeCapability{
				utils.NewMountCapability(0, fsType, mntFlags...),
			},
			// ControllerCreateCredentials: userCreds,
			Parameters: params,
		}
		res, err := client.CreateVolume(ctx, req)
		if res == nil {
			return nil, err
		}
		return res.Volume, err
	}

	createNewSnapshotWithResult := func() (*csi.Snapshot, error) {
		req := &csi.CreateSnapshotRequest{
			Name:           snapName,
			SourceVolumeId: volID,
			Parameters:     params,
		}
		res, err := client.CreateSnapshot(ctx, req)
		if res == nil {
			return nil, err
		}
		return res.Snapshot, err
	}

	expandVolumeWithResult := func() (int64, error) {
		req := &csi.ControllerExpandVolumeRequest{
			VolumeId: volID,
			CapacityRange: &csi.CapacityRange{
				RequiredBytes: limBytes,
				LimitBytes:    expBytes,
			},
		}
		res, err := client.ControllerExpandVolume(ctx, req)
		if res == nil {
			return 0, err
		}
		return res.CapacityBytes, err
	}

	createNewSnapshot := func() {
		snap, err = createNewSnapshotWithResult()
	}

	createNewVolume := func() {
		vol, err = createNewVolumeWithResult()
	}

	expandVolume := func() {
		capBytes, err = expandVolumeWithResult()
	}

	validateNewVolumeResult := func(
		vol *csi.Volume,
		err error) bool {

		if err != nil {
			Ω(err).Should(ΣCM(codes.Aborted, "pending"))
			return true
		}

		Ω(vol).ShouldNot(BeNil())
		Ω(vol.CapacityBytes).Should(Equal(limBytes))
		Ω(vol.VolumeId).Should(Equal(volID))
		Ω(vol.VolumeContext["name"]).Should(Equal(volName))
		return false
	}

	validateNewSnapshotResult := func(
		snap *csi.Snapshot,
		err error) bool {

		if err != nil {
			Ω(err).Should(ΣCM(codes.Aborted, "pending"))
			return true
		}

		Ω(snap).ShouldNot(BeNil())
		Ω(snap.SnapshotId).Should(Equal(snapID))
		Ω(snap.SourceVolumeId).Should(Equal(volID))
		return false
	}

	validateVolumeExpandResult := func(
		bytes int64,
		err error) bool {

		if err != nil {
			Ω(err).Should(ΣCM(codes.Aborted, "pending"))
			return true
		}

		Ω(bytes).Should(Equal(expBytes))
		return false
	}

	validateNewSnapshot := func() {
		validateNewSnapshotResult(snap, err)
	}

	validateNewVolume := func() {
		validateNewVolumeResult(vol, err)
	}

	validateVolumeExpand := func() {
		validateVolumeExpandResult(capBytes, err)
	}

	Describe("CreateSnapshot", func() {
		JustBeforeEach(func() {
			vol, err = createNewVolumeWithResult()
			createNewSnapshot()
		})
		Context("Normal Create Volume Call", func() {
			It("Should Be Valid", validateNewSnapshot)
		})
	})

	Describe("CreateVolume", func() {
		JustBeforeEach(func() {
			createNewVolume()
		})
		Context("Normal Create Volume Call", func() {
			It("Should Be Valid", validateNewVolume)
		})
		Context("Field Size Error", func() {
			Context("Invalid Name", func() {
				BeforeEach(func() {
					volName = string129
				})
				It("Should Be Invalid", func() {
					Ω(err).Should(HaveOccurred())
					Ω(vol).Should(BeNil())
					Ω(err).Should(ΣCM(
						codes.InvalidArgument,
						"exceeds size limit: Name: max=128, size=129"))
				})
			})
			Context("Invalid Params Field Key", func() {
				BeforeEach(func() {
					params[string129] = "class"
				})
				It("Should Be Invalid", func() {
					Ω(err).Should(HaveOccurred())
					Ω(vol).Should(BeNil())
					Ω(err).Should(ΣCM(
						codes.InvalidArgument,
						fmt.Sprintf(
							"exceeds size limit: Parameters[%s]: max=128, size=129",
							string129)))
				})
			})
			Context("Invalid Params Field Val", func() {
				BeforeEach(func() {
					params["class"] = string129
				})
				It("Should Be Invalid", func() {
					Ω(err).Should(HaveOccurred())
					Ω(vol).Should(BeNil())
					Ω(err).Should(ΣCM(
						codes.InvalidArgument,
						"exceeds size limit: Parameters[class]=: max=128, size=129"))
				})
			})
			Context("Invalid Params Map", func() {
				BeforeEach(func() {
					for i := 0; i < 48; i++ {
						params[fmt.Sprintf("%d", i)] = string128
					}
				})
				It("Should Be Invalid", func() {
					Ω(err).Should(HaveOccurred())
					Ω(vol).Should(BeNil())
					Ω(err).Should(ΣCM(
						codes.InvalidArgument,
						"exceeds size limit: Parameters: max=4096, size=6237"))
				})
			})
		})
		Context("No LimitBytes", func() {
			BeforeEach(func() {
				limBytes = 0
			})
			It("Should Be Valid", func() {
				Ω(err).ShouldNot(HaveOccurred())
				Ω(vol).ShouldNot(BeNil())
				Ω(vol.CapacityBytes).Should(Equal(reqBytes))
				Ω(vol.VolumeContext["name"]).Should(Equal(volName))
			})
		})
		Context("Missing Name", func() {
			BeforeEach(func() {
				volName = ""
			})
			It("Should Be Invalid", func() {
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(ΣCM(codes.InvalidArgument, "required: Name"))
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
				Ω(err).Should(ΣCM(codes.InvalidArgument, "required: VolumeID"))
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
		})
	})

	Describe("DeleteSnapshot", func() {
		var snapID string
		BeforeEach(func() {
			snapID = CTest().ComponentTexts[2]
		})
		AfterEach(func() {
			snapID = ""
		})
		deleteSnapshot := func() {
			_, err = client.DeleteSnapshot(
				ctx,
				&csi.DeleteSnapshotRequest{
					SnapshotId: snapID,
				})
		}
		JustBeforeEach(func() {
			deleteSnapshot()
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
		Context("Missing Snapshot ID", func() {
			BeforeEach(func() {
				snapID = ""
			})
			It("Should Not Be Valid", func() {
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(ΣCM(codes.InvalidArgument, "required: SnapshotID"))
			})
		})
		Context("Not Found", func() {
			BeforeEach(func() {
				snapID = "5"
			})

			It("Should Be Valid", func() {
				Ω(err).ShouldNot(HaveOccurred())
				deleteSnapshot()
				Ω(err).ShouldNot(HaveOccurred())
				deleteSnapshot()
				Ω(err).ShouldNot(HaveOccurred())
				deleteSnapshot()
				Ω(err).ShouldNot(HaveOccurred())
			})
		})
	})

	Describe("ListVolumes", func() {
		var vols []csi.Volume
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

	Describe("ListSnapshots", func() {
		var snaps []csi.Snapshot
		AfterEach(func() {
			snaps = nil
		})
		JustBeforeEach(func() {
			snaps, err = listSnapshots()
		})
		Context("Normal List Snapshots Call", func() {
			It("Should Be Valid", func() {
				Ω(err).ShouldNot(HaveOccurred())
				Ω(snaps).Should(BeNil())
			})
		})
		Context("Create five Snapshots Then List", func() {
			JustBeforeEach(func() {
				for i := 0; i < 5; i++ {
					createNewSnapshot()
					validateNewSnapshot()
				}
				snaps, err = listSnapshots()
			})
			It("Should Be Valid", func() {
				Ω(err).ShouldNot(HaveOccurred())
				Ω(snaps).ShouldNot(BeNil())
				Ω(snaps).Should(HaveLen(5))
			})
		})
	})

	Describe("Publication", func() {

		var (
			res *csi.ControllerPublishVolumeResponse
			err error
		)

		devPathKey := path.Join(service.Name, "dev")

		publishVolume := func() {
			req := &csi.ControllerPublishVolumeRequest{
				VolumeId: "1",
				NodeId:   nodeId,
				Readonly: true,
				VolumeCapability: utils.NewMountCapability(
					csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					"mock"),
			}
			res, err = client.ControllerPublishVolume(ctx, req)
		}

		shouldBePublished := func() {
			Ω(err).ShouldNot(HaveOccurred())
			pubInfo = res.PublishContext
			Ω(err).ShouldNot(HaveOccurred())
			Ω(pubInfo).ShouldNot(BeNil())
			Ω(pubInfo["device"]).Should(Equal("/dev/mock"))
		}

		unpublishVolume := func() {
			_, err := client.ControllerUnpublishVolume(
				ctx,
				&csi.ControllerUnpublishVolumeRequest{
					VolumeId: "1",
					NodeId:   nodeId,
				})
			Ω(err).ShouldNot(HaveOccurred())
		}

		JustBeforeEach(func() {
			publishVolume()
		})

		Context("PublishVolume", func() {

			Context("Valid Node ID", func() {
				It("Should Be Valid", func() {
					shouldBePublished()
					vols, err := listVolumes()
					Ω(err).ShouldNot(HaveOccurred())
					Ω(vols).Should(HaveLen(3))
					Ω(vols[0].VolumeContext[devPathKey]).Should(Equal("/dev/mock"))
				})
			})

			Context("Invalid Node ID", func() {

				BeforeEach(func() {
					nodeId = ""
				})

				It("Should Error", func() {
					Ω(err).Should(HaveOccurred())
					Ω(err).Should(ΣCM(codes.InvalidArgument, "required: NodeID"))
				})
			})

		})

		Context("UnpublishVolume", func() {
			JustBeforeEach(func() {
				shouldBePublished()
			})

			Context("Valid NodeID", func() {
				It("Should Be Unpublished", func() {
					unpublishVolume()
					vols, err := listVolumes()
					Ω(err).ShouldNot(HaveOccurred())
					Ω(vols).Should(HaveLen(3))
					_, ok := vols[0].VolumeContext[devPathKey]
					Ω(ok).Should(BeFalse())
				})
			})

			Context("No NodeID", func() {
				It("Should Still Be Unpublished", func() {
					nodeId = ""
					unpublishVolume()
					vols, err := listVolumes()
					Ω(err).ShouldNot(HaveOccurred())
					Ω(vols).Should(HaveLen(3))
					_, ok := vols[0].VolumeContext[devPathKey]
					Ω(ok).Should(BeFalse())
				})
			})
		})
	})

	Describe("Expand Volume", func() {
		JustBeforeEach(func() {
			createNewVolume()
			expandVolume()
		})
		Context("ExpandVolume", func() {
			It("Should be expanded", validateVolumeExpand)
		})
	})
})
