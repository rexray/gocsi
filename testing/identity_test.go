package gocsi_test

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/dell/gocsi"
	csictx "github.com/dell/gocsi/context"
	"github.com/dell/gocsi/mock/service"
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
			// reqVersion    string
		)
		JustBeforeEach(func() {
			var res *csi.GetPluginInfoResponse
			res, err = client.GetPluginInfo(ctx, &csi.GetPluginInfoRequest{})
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
		It("Should be valid", func() {
			Ω(err).ShouldNot(HaveOccurred())
			Ω(name).Should(Equal(service.Name))
			Ω(vendorVersion).Should(Equal(service.VendorVersion))
			Ω(manifest).Should(HaveLen(1))
			Ω(manifest["url"]).Should(Equal(service.Manifest["url"]))
		})

		Context("With Invalid Plug-in Name Error", func() {
			BeforeEach(func() {
				// reqVersion = "0.2.0"
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

	Describe("GetPluginCapabilities", func() {
		It("Should Be Valid", func() {
			rep, err := client.GetPluginCapabilities(
				ctx, &csi.GetPluginCapabilitiesRequest{})
			Ω(err).ShouldNot(HaveOccurred())
			Ω(rep).ShouldNot(BeNil())
			Ω(rep.Capabilities).Should(HaveLen(2))
			svc := rep.Capabilities[0].GetService()
			Ω(svc).ShouldNot(BeNil())
			Ω(svc.Type).Should(Equal(csi.PluginCapability_Service_CONTROLLER_SERVICE))
			svc2 := rep.Capabilities[1].GetVolumeExpansion()
			Ω(svc2).ShouldNot(BeNil())
			Ω(svc2.Type).Should(Equal(csi.PluginCapability_VolumeExpansion_ONLINE))
		})
	})

	Describe("Probe", func() {
		It("Should Be Ready", func() {
			rep, err := client.Probe(
				ctx, &csi.ProbeRequest{})
			Ω(err).ShouldNot(HaveOccurred())
			Ω(rep).ShouldNot(BeNil())
			Ω(rep.GetReady().GetValue()).To(Equal(true))
		})
	})
})
