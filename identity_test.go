package gocsi_test

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/thecodeteam/gocsi"
	"github.com/thecodeteam/gocsi/mock/service"
)

var _ = Describe("Identity", func() {
	var (
		err      error
		stopMock func()
		ctx      context.Context
		gclient  *grpc.ClientConn
		client   csi.IdentityClient
	)
	BeforeEach(func() {
		ctx = context.Background()
		gclient, stopMock, err = startMockServer(ctx)
		Ω(err).ShouldNot(HaveOccurred())
		client = csi.NewIdentityClient(gclient)
	})
	AfterEach(func() {
		ctx = nil
		gclient.Close()
		gclient = nil
		client = nil
		stopMock()
	})

	Describe("GetPluginInfo", func() {
		var (
			name          string
			vendorVersion string
			manifest      map[string]string
			version       csi.Version
		)
		BeforeEach(func() {
			var ok bool
			version, ok = gocsi.ParseVersion(CTest().ComponentTexts[3])
			Ω(ok).ShouldNot(BeFalse())
			var res *csi.GetPluginInfoResponse
			res, err = client.GetPluginInfo(ctx, &csi.GetPluginInfoRequest{
				Version: &csi.Version{
					Major: version.GetMajor(),
					Minor: version.GetMinor(),
					Patch: version.GetPatch(),
				},
			})
			if err == nil {
				name = res.Name
				vendorVersion = res.VendorVersion
				manifest = res.Manifest
			}
		})
		AfterEach(func() {
			name = ""
			vendorVersion = ""
			manifest = nil
		})
		shouldBeValid := func() {
			Ω(err).ShouldNot(HaveOccurred())
			Ω(name).Should(Equal(service.Name))
			Ω(vendorVersion).Should(Equal(service.VendorVersion))
			Ω(manifest).Should(HaveLen(1))
			Ω(manifest["url"]).Should(Equal(service.Manifest["url"]))
		}
		shouldNotBeValid := func() {
			Ω(err).Should(ΣCM(
				codes.InvalidArgument,
				fmt.Sprintf("invalid request version: %s",
					CTest().ComponentTexts[3])))

		}
		Context("With Request Version", func() {
			Context("0.0.0", func() {
				It("Should Not Be Valid", shouldNotBeValid)
			})
			Context("0.1.0", func() {
				It("Should Be Valid", shouldBeValid)
			})
			Context("0.2.0", func() {
				It("Should Be Valid", shouldBeValid)
			})
			Context("1.0.0", func() {
				It("Should Be Valid", shouldBeValid)
			})
			Context("1.1.0", func() {
				It("Should Be Valid", shouldBeValid)
			})
			Context("1.2.0", func() {
				It("Should Not Be Valid", shouldNotBeValid)
			})
		})
	})

	Describe("GetSupportedVersions", func() {
		It("Should Be Valid", func() {
			rep, err := client.GetSupportedVersions(
				ctx, &csi.GetSupportedVersionsRequest{})
			Ω(err).ShouldNot(HaveOccurred())
			Ω(rep).ShouldNot(BeNil())
			fmt.Fprintf(
				GinkgoWriter, "expVersions: %v\n", mockSupportedVersions)
			fmt.Fprintf(
				GinkgoWriter, "actVersions: %v\n", rep.SupportedVersions)

			Ω(rep.SupportedVersions).Should(HaveLen(len(mockSupportedVersions)))
			for i, v := range rep.SupportedVersions {
				Ω(*v).Should(Equal(mockSupportedVersions[i]))
			}
		})
	})
})
