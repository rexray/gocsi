package gocsi_test

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/thecodeteam/gocsi"
	csictx "github.com/thecodeteam/gocsi/context"
	"github.com/thecodeteam/gocsi/mock/service"
	"github.com/thecodeteam/gocsi/utils"
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
	})
	JustBeforeEach(func() {
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
			reqVersion    string
		)
		JustBeforeEach(func() {
			var ok bool
			version, ok = utils.ParseVersion(reqVersion)
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
				fmt.Sprintf("invalid: Version=%s", CTest().ComponentTexts[3])))
		}

		Context("With Request Version", func() {
			BeforeEach(func() {
				reqVersion = CTest().ComponentTexts[3]
			})
			Context("0.0.0", func() {
				It("Should Not Be Valid", shouldNotBeValid)
			})
			Context("0.1.0", func() {
				It("Should Be Valid", shouldNotBeValid)
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

		Context("With Invalid Plug-in Name Error", func() {
			BeforeEach(func() {
				reqVersion = "0.2.0"
				ctx = csictx.WithEnviron(ctx,
					[]string{
						gocsi.EnvVarPluginInfo + "=Mock,v1.0.0",
					})
			})
			It("Should Not Be Valid", func() {
				Ω(err).Should(ΣCM(
					codes.Internal,
					"invalid: Name=Mock: patt=%s",
					`^[\w\d]+\.[\w\d\.\-_]*[\w\d]$`))
				st, ok := status.FromError(err)
				Ω(ok).Should(BeTrue())
				Ω(st.Details()).Should(HaveLen(1))
				rep, ok := st.Details()[0].(*csi.GetPluginInfoResponse)
				Ω(ok).Should(BeTrue())
				Ω(rep.Name).Should(Equal("Mock"))
				Ω(rep.VendorVersion).Should(Equal("v1.0.0"))
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

	Describe("GetPluginCapabilities", func() {
		It("Should Be Valid", func() {
			rep, err := client.GetPluginCapabilities(
				ctx, &csi.GetPluginCapabilitiesRequest{
					Version: &csi.Version{
						Major: 0,
						Minor: 2,
						Patch: 0,
					},
				})
			Ω(err).ShouldNot(HaveOccurred())
			Ω(rep).ShouldNot(BeNil())
			Ω(rep.Capabilities).Should(HaveLen(1))
			svc := rep.Capabilities[0].GetService()
			Ω(svc).ShouldNot(BeNil())
			Ω(svc.Type).Should(Equal(csi.PluginCapability_Service_CONTROLLER_SERVICE))
		})
	})
})
