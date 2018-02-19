package specvalidator

import (
	"reflect"
	"regexp"
	"sync"

	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/container-storage-interface/spec/lib/go/csi"

	"github.com/rexray/gocsi/utils"
)

// Option configures the spec validator interceptor.
type Option func(*opts)

type opts struct {
	sync.Mutex
	supportedVersions   []csi.Version
	reqValidation       bool
	repValidation       bool
	requiresNodeID      bool
	requiresPubVolInfo  bool
	requiresVolAttribs  bool
	requiresCredentials map[string]struct{}
}

func (o *opts) requireCredentials(m string) {
	o.Lock()
	defer o.Unlock()
	if o.requiresCredentials == nil {
		o.requiresCredentials = map[string]struct{}{}
	}
	o.requiresCredentials[m] = struct{}{}
}

// WithSupportedVersions is a Option that indicates the
// list of versions supported by any CSI RPC that participates in
// version validation.
func WithSupportedVersions(versions ...csi.Version) Option {
	return func(o *opts) {
		o.supportedVersions = versions
	}
}

// WithRequestValidation is a Option that enables request validation.
func WithRequestValidation() Option {
	return func(o *opts) {
		o.reqValidation = true
	}
}

// WithResponseValidation is a Option that enables response validation.
func WithResponseValidation() Option {
	return func(o *opts) {
		o.repValidation = true
	}
}

// WithRequiresNodeID is a Option that indicates
// ControllerPublishVolume requests and GetNodeID responses must
// contain non-empty node ID data.
func WithRequiresNodeID() Option {
	return func(o *opts) {
		o.requiresNodeID = true
	}
}

// WithRequiresPublishVolumeInfo is a Option that indicates
// ControllerPublishVolume responses and NodePublishVolume requests must
// contain non-empty publish volume info data.
func WithRequiresPublishVolumeInfo() Option {
	return func(o *opts) {
		o.requiresPubVolInfo = true
	}
}

// WithRequiresVolumeAttributes is a Option that indicates
// ControllerPublishVolume, ValidateVolumeCapabilities, and NodePublishVolume
// requests must contain non-empty volume attribute data.
func WithRequiresVolumeAttributes() Option {
	return func(o *opts) {
		o.requiresVolAttribs = true
	}
}

// WithRequiresCreateVolumeCredentials is a Option
// that indicates the eponymous requests must contain non-empty credentials
// data.
func WithRequiresCreateVolumeCredentials() Option {
	return func(o *opts) {
		o.requireCredentials(utils.CreateVolume)
	}
}

// WithRequiresDeleteVolumeCredentials is a Option
// that indicates the eponymous requests must contain non-empty credentials
// data.
func WithRequiresDeleteVolumeCredentials() Option {
	return func(o *opts) {
		o.requireCredentials(utils.DeleteVolume)
	}
}

// WithRequiresControllerPublishVolumeCredentials is a Option
// that indicates the eponymous requests must contain non-empty credentials
// data.
func WithRequiresControllerPublishVolumeCredentials() Option {
	return func(o *opts) {
		o.requireCredentials(utils.ControllerPublishVolume)
	}
}

// WithRequiresControllerUnpublishVolumeCredentials is a Option
// that indicates the eponymous requests must contain non-empty credentials
// data.
func WithRequiresControllerUnpublishVolumeCredentials() Option {
	return func(o *opts) {
		o.requireCredentials(utils.ControllerUnpublishVolume)
	}
}

// WithRequiresNodePublishVolumeCredentials is a Option
// that indicates the eponymous requests must contain non-empty credentials
// data.
func WithRequiresNodePublishVolumeCredentials() Option {
	return func(o *opts) {
		o.requireCredentials(utils.NodePublishVolume)
	}
}

// WithRequiresNodeUnpublishVolumeCredentials is a Option
// that indicates the eponymous requests must contain non-empty credentials
// data.
func WithRequiresNodeUnpublishVolumeCredentials() Option {
	return func(o *opts) {
		o.requireCredentials(utils.NodeUnpublishVolume)
	}
}

type interceptor struct {
	opts opts
}

// NewServerSpecValidator returns a new UnaryServerInterceptor that validates
// server request and response data against the CSI specification.
func NewServerSpecValidator(
	opts ...Option) grpc.UnaryServerInterceptor {

	return newSpecValidator(opts...).handleServer
}

// NewClientSpecValidator provides a UnaryClientInterceptor that validates
// client request and response data against the CSI specification.
func NewClientSpecValidator(
	opts ...Option) grpc.UnaryClientInterceptor {

	return newSpecValidator(opts...).handleClient
}

func newSpecValidator(opts ...Option) *interceptor {
	i := &interceptor{}
	for _, withOpts := range opts {
		withOpts(&i.opts)
	}
	return i
}

func (s *interceptor) handleServer(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	return s.handle(ctx, info.FullMethod, req, func() (interface{}, error) {
		return handler(ctx, req)
	})
}

func (s *interceptor) handleClient(
	ctx context.Context,
	method string,
	req, rep interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption) error {

	_, err := s.handle(ctx, method, req, func() (interface{}, error) {
		return rep, invoker(ctx, method, req, rep, cc, opts...)
	})
	return err
}

func (s *interceptor) handle(
	ctx context.Context,
	method string,
	req interface{},
	next func() (interface{}, error)) (interface{}, error) {

	// If the request is nil then pass control to the next handler
	// in the chain.
	if req == nil {
		return next()
	}

	if s.opts.reqValidation {
		// Validate the request version.
		if err := s.validateRequestVersion(ctx, req); err != nil {
			return nil, err
		}

		// Validate the request against the CSI specification.
		if err := s.validateRequest(ctx, method, req); err != nil {
			return nil, err
		}
	}

	// Use the function passed into this one to get the response. On the
	// server-side this could possibly invoke additional interceptors or
	// the RPC. On the client side this invokes the RPC.
	rep, err := next()

	if err != nil {
		return nil, err
	}

	// Determine whether or not the response is nil. Otherwise it
	// will no longer be possible to perform a nil equality check on the
	// response to the interface{} rules for nil comparison.
	//
	// If the response is nil then go ahead and return a nil value
	// directly in order to fulfill Go's rules about nil values and
	// interface{} types.
	if utils.IsNilResponse(method, rep) {
		return nil, nil
	}

	log.WithField(
		"repValidation", s.opts.repValidation).Debug("do rep validtion?")
	if s.opts.repValidation {
		// Validate the response against the CSI specification.
		if err := s.validateResponse(ctx, method, rep); err != nil {

			// If an error occurred while validating the response, it is
			// imperative the response not be discarded as it could be
			// important to the client.
			st, ok := status.FromError(err)
			if !ok {
				st = status.New(codes.Internal, err.Error())
			}

			// Add the response to the error details.
			st, err2 := st.WithDetails(rep.(proto.Message))

			// If there is a problem encoding the response into the
			// protobuf details then err on the side of caution, log
			// the encoding error, validation error, and return the
			// original response.
			if err2 != nil {
				log.WithFields(map[string]interface{}{
					"encErr": err2,
					"valErr": err,
				}).Error("failed to encode error details; " +
					"returning invalid response")

				return rep, nil
			}

			// There was no issue encoding the response, so return
			// the gRPC status error with the error message and payload.
			return nil, st.Err()
		}
	}

	return rep, err
}

type interceptorHasVolumeID interface {
	GetVolumeId() string
}
type interceptorHasNodeID interface {
	GetNodeId() string
}
type interceptorHasUserCredentials interface {
	GetUserCredentials() map[string]string
}
type interceptorHasVolumeAttributes interface {
	GetVolumeAttributes() map[string]string
}
type interceptorHasVersion interface {
	GetVersion() *csi.Version
}

func (s *interceptor) validateRequest(
	ctx context.Context,
	method string,
	req interface{}) error {

	if req == nil {
		return nil
	}

	// Validate field sizes.
	if err := validateFieldSizes(req); err != nil {
		return err
	}

	// Check to see if the request has a volume ID and if it is set.
	// If the volume ID is not set then return an error.
	if treq, ok := req.(interceptorHasVolumeID); ok {
		if treq.GetVolumeId() == "" {
			return status.Error(
				codes.InvalidArgument, "required: VolumeID")
		}
	}

	// Check to see if the request has a node ID and if it is set.
	// If the node ID is not set then return an error.
	if s.opts.requiresNodeID {
		if treq, ok := req.(interceptorHasNodeID); ok {
			if treq.GetNodeId() == "" {
				return status.Error(
					codes.InvalidArgument, "required: NodeID")
			}
		}
	}

	// Check to see if the request has credentials and if they're required.
	// If the credentials are required but no credentials are specified then
	// return an error.
	if _, ok := s.opts.requiresCredentials[method]; ok {
		if treq, ok := req.(interceptorHasUserCredentials); ok {
			if len(treq.GetUserCredentials()) == 0 {
				return status.Error(
					codes.InvalidArgument, "required: UserCredentials")
			}
		}
	}

	// Check to see if the request has volume attributes and if they're
	// required. If the volume attributes are required by no attributes are
	// specified then return an error.
	if s.opts.requiresVolAttribs {
		if treq, ok := req.(interceptorHasVolumeAttributes); ok {
			if len(treq.GetVolumeAttributes()) == 0 {
				return status.Error(
					codes.InvalidArgument, "required: VolumeAttributes")
			}
		}
	}

	// Please leave requests that do not require explicit validation commented
	// out for purposes of optimization. These requests are retained in this
	// form to make it easy to add validation later if required.
	//
	switch tobj := req.(type) {
	//
	// Controller Service
	//
	case *csi.CreateVolumeRequest:
		return s.validateCreateVolumeRequest(ctx, *tobj)
	case *csi.ControllerPublishVolumeRequest:
		return s.validateControllerPublishVolumeRequest(ctx, *tobj)
	case *csi.ValidateVolumeCapabilitiesRequest:
		return s.validateValidateVolumeCapabilitiesRequest(ctx, *tobj)
	case *csi.GetCapacityRequest:
		return s.validateGetCapacityRequest(ctx, *tobj)
	//
	// Node Service
	//
	case *csi.NodePublishVolumeRequest:
		return s.validateNodePublishVolumeRequest(ctx, *tobj)
	case *csi.NodeUnpublishVolumeRequest:
		return s.validateNodeUnpublishVolumeRequest(ctx, *tobj)
	}

	return nil
}

func (s *interceptor) validateResponse(
	ctx context.Context,
	method string,
	rep interface{}) error {

	if utils.IsNilResponse(method, rep) {
		return nil
	}

	// Validate the field sizes.
	if err := validateFieldSizes(rep); err != nil {
		return err
	}

	switch tobj := rep.(type) {
	//
	// Controller Service
	//
	case *csi.CreateVolumeResponse:
		return s.validateCreateVolumeResponse(ctx, *tobj)
	case *csi.ControllerPublishVolumeResponse:
		return s.validateControllerPublishVolumeResponse(ctx, *tobj)
	case *csi.ListVolumesResponse:
		return s.validateListVolumesResponse(ctx, *tobj)
	case *csi.ControllerGetCapabilitiesResponse:
		return s.validateControllerGetCapabilitiesResponse(ctx, *tobj)
	//
	// Identity Service
	//
	case *csi.GetSupportedVersionsResponse:
		return s.validateGetSupportedVersionsResponse(ctx, *tobj)
	case *csi.GetPluginInfoResponse:
		return s.validateGetPluginInfoResponse(ctx, *tobj)
	//
	// Node Service
	//
	case *csi.GetNodeIDResponse:
		return s.validateGetNodeIDResponse(ctx, *tobj)
	case *csi.NodeGetCapabilitiesResponse:
		return s.validateNodeGetCapabilitiesResponse(ctx, *tobj)
	}

	return nil
}

func (s *interceptor) validateRequestVersion(
	ctx context.Context,
	req interface{}) error {

	// Check to see if the request version should be validated.
	if len(s.opts.supportedVersions) == 0 {
		return nil
	}

	treq, ok := req.(interceptorHasVersion)
	if !ok {
		return nil
	}

	var (
		supported      bool
		requestVersion = treq.GetVersion()
	)

	if requestVersion == nil {
		return status.Error(codes.InvalidArgument, "nil: Version")
	}

	for _, supportedVersion := range s.opts.supportedVersions {
		if utils.CompareVersions(requestVersion, &supportedVersion) == 0 {
			supported = true
			break
		}
	}

	if !supported {
		return status.Errorf(
			codes.InvalidArgument,
			"invalid: Version=%s",
			utils.SprintfVersion(*requestVersion))
	}

	return nil
}

func (s *interceptor) validateCreateVolumeRequest(
	ctx context.Context,
	req csi.CreateVolumeRequest) error {

	if req.Name == "" {
		return status.Error(
			codes.InvalidArgument, "required: Name")
	}

	return validateVolumeCapabilitiesArg(req.VolumeCapabilities, true)
}

// func (s *interceptor) validateDeleteVolumeRequest(
// 	ctx context.Context,
// 	req csi.DeleteVolumeRequest) error {
//
// 	return nil
// }

func (s *interceptor) validateControllerPublishVolumeRequest(
	ctx context.Context,
	req csi.ControllerPublishVolumeRequest) error {

	return validateVolumeCapabilityArg(req.VolumeCapability, true)
}

func (s *interceptor) validateValidateVolumeCapabilitiesRequest(
	ctx context.Context,
	req csi.ValidateVolumeCapabilitiesRequest) error {

	return validateVolumeCapabilitiesArg(req.VolumeCapabilities, true)
}

func (s *interceptor) validateGetCapacityRequest(
	ctx context.Context,
	req csi.GetCapacityRequest) error {

	return validateVolumeCapabilitiesArg(req.VolumeCapabilities, false)
}

func (s *interceptor) validateNodePublishVolumeRequest(
	ctx context.Context,
	req csi.NodePublishVolumeRequest) error {

	if req.TargetPath == "" {
		return status.Error(
			codes.InvalidArgument, "required: TargetPath")
	}

	if s.opts.requiresPubVolInfo && len(req.PublishVolumeInfo) == 0 {
		return status.Error(
			codes.InvalidArgument, "required: PublishVolumeInfo")
	}

	return validateVolumeCapabilityArg(req.VolumeCapability, true)
}

func (s *interceptor) validateNodeUnpublishVolumeRequest(
	ctx context.Context,
	req csi.NodeUnpublishVolumeRequest) error {

	if req.TargetPath == "" {
		return status.Error(
			codes.InvalidArgument, "required: TargetPath")
	}

	return nil
}

func (s *interceptor) validateCreateVolumeResponse(
	ctx context.Context,
	rep csi.CreateVolumeResponse) error {

	if rep.VolumeInfo == nil {
		return status.Error(codes.Internal, "nil: VolumeInfo")
	}

	if rep.VolumeInfo.Id == "" {
		return status.Error(codes.Internal, "empty: VolumeInfo.Id")
	}

	if s.opts.requiresVolAttribs && len(rep.VolumeInfo.Attributes) == 0 {
		return status.Error(
			codes.Internal, "non-nil, empty: VolumeInfo.Attributes")
	}

	return nil
}

func (s *interceptor) validateControllerPublishVolumeResponse(
	ctx context.Context,
	rep csi.ControllerPublishVolumeResponse) error {

	if s.opts.requiresPubVolInfo && len(rep.PublishVolumeInfo) == 0 {
		return status.Error(codes.Internal, "empty: PublishVolumeInfo")
	}
	return nil
}

func (s *interceptor) validateListVolumesResponse(
	ctx context.Context,
	rep csi.ListVolumesResponse) error {

	for i, e := range rep.Entries {
		volInfo := e.VolumeInfo
		if volInfo == nil {
			return status.Errorf(
				codes.Internal,
				"nil: Entries[%d].VolumeInfo", i)
		}
		if volInfo.Id == "" {
			return status.Errorf(
				codes.Internal,
				"empty: Entries[%d].VolumeInfo.Id", i)
		}
		if volInfo.Attributes != nil && len(volInfo.Attributes) == 0 {
			return status.Errorf(
				codes.Internal,
				"non-nil, empty: Entries[%d].VolumeInfo.Attributes", i)
		}
	}

	return nil
}

func (s *interceptor) validateControllerGetCapabilitiesResponse(
	ctx context.Context,
	rep csi.ControllerGetCapabilitiesResponse) error {

	if rep.Capabilities != nil && len(rep.Capabilities) == 0 {
		return status.Error(codes.Internal, "non-nil, empty: Capabilities")
	}
	return nil
}

func (s *interceptor) validateGetSupportedVersionsResponse(
	ctx context.Context,
	rep csi.GetSupportedVersionsResponse) error {

	if len(rep.SupportedVersions) == 0 {
		return status.Error(codes.Internal, "empty: SupportedVersions")
	}
	return nil
}

const (
	pluginNameMax           = 63
	pluginNamePatt          = `^[\w\d]+\.[\w\d\.\-_]*[\w\d]$`
	pluginVendorVersionPatt = `^v?(\d+\.){2}(\d+)(-.+)?$`
)

func (s *interceptor) validateGetPluginInfoResponse(
	ctx context.Context,
	rep csi.GetPluginInfoResponse) error {

	log.Debug("validateGetPluginInfoResponse: enter")

	if rep.Name == "" {
		return status.Error(codes.Internal, "empty: Name")
	}
	if l := len(rep.Name); l > pluginNameMax {
		return status.Errorf(codes.Internal,
			"exceeds size limit: Name=%s: max=%d, size=%d",
			rep.Name, pluginNameMax, l)
	}
	nok, err := regexp.MatchString(pluginNamePatt, rep.Name)
	if err != nil {
		return err
	}
	if !nok {
		return status.Errorf(codes.Internal,
			"invalid: Name=%s: patt=%s",
			rep.Name, pluginNamePatt)
	}
	if rep.VendorVersion == "" {
		return status.Error(codes.Internal, "empty: VendorVersion")
	}
	vok, err := regexp.MatchString(pluginVendorVersionPatt, rep.VendorVersion)
	if err != nil {
		return err
	}
	if !vok {
		return status.Errorf(codes.Internal,
			"invalid: VendorVersion=%s: patt=%s",
			rep.VendorVersion, pluginVendorVersionPatt)
	}
	if rep.Manifest != nil && len(rep.Manifest) == 0 {
		return status.Error(codes.Internal,
			"non-nil, empty: Manifest")
	}
	return nil
}

func (s *interceptor) validateGetNodeIDResponse(
	ctx context.Context,
	rep csi.GetNodeIDResponse) error {

	if s.opts.requiresNodeID && rep.NodeId == "" {
		return status.Error(codes.Internal, "empty: NodeID")
	}
	return nil
}

func (s *interceptor) validateNodeGetCapabilitiesResponse(
	ctx context.Context,
	rep csi.NodeGetCapabilitiesResponse) error {

	if rep.Capabilities != nil && len(rep.Capabilities) == 0 {
		return status.Error(codes.Internal, "non-nil, empty: Capabilities")
	}
	return nil
}

func validateVolumeCapabilityArg(
	volCap *csi.VolumeCapability,
	required bool) error {

	if required && volCap == nil {
		return status.Error(codes.InvalidArgument, "required: VolumeCapability")
	}

	if volCap.AccessMode == nil {
		return status.Error(codes.InvalidArgument, "required: AccessMode")
	}

	atype := volCap.GetAccessType()
	if atype == nil {
		return status.Error(codes.InvalidArgument, "required: AccessType")
	}

	switch tatype := atype.(type) {
	case *csi.VolumeCapability_Block:
		if tatype.Block == nil {
			return status.Error(codes.InvalidArgument,
				"required: AccessType.Block")
		}
	case *csi.VolumeCapability_Mount:
		if tatype.Mount == nil {
			return status.Error(codes.InvalidArgument,
				"required: AccessType.Mount")
		}
	default:
		return status.Errorf(codes.InvalidArgument,
			"invalid: AccessType=%T", atype)
	}

	return nil
}

func validateVolumeCapabilitiesArg(
	volCaps []*csi.VolumeCapability,
	required bool) error {

	if len(volCaps) == 0 {
		if required {
			return status.Error(
				codes.InvalidArgument, "required: VolumeCapabilities")
		}
		return nil
	}

	for i, cap := range volCaps {
		if cap.AccessMode == nil {
			return status.Errorf(
				codes.InvalidArgument,
				"required: VolumeCapabilities[%d].AccessMode", i)
		}
		atype := cap.GetAccessType()
		if atype == nil {
			return status.Errorf(
				codes.InvalidArgument,
				"required: VolumeCapabilities[%d].AccessType", i)
		}
		switch tatype := atype.(type) {
		case *csi.VolumeCapability_Block:
			if tatype.Block == nil {
				return status.Errorf(
					codes.InvalidArgument,
					"required: VolumeCapabilities[%d].AccessType.Block", i)

			}
		case *csi.VolumeCapability_Mount:
			if tatype.Mount == nil {
				return status.Errorf(
					codes.InvalidArgument,
					"required: VolumeCapabilities[%d].AccessType.Mount", i)
			}
		default:
			return status.Errorf(
				codes.InvalidArgument,
				"invalid: VolumeCapabilities[%d].AccessType=%T", i, atype)
		}
	}

	return nil
}

const (
	maxFieldString = 128
	maxFieldMap    = 4096
)

func validateFieldSizes(msg interface{}) error {
	rv := reflect.ValueOf(msg).Elem()
	tv := rv.Type()
	nf := tv.NumField()
	for i := 0; i < nf; i++ {
		f := rv.Field(i)
		switch f.Kind() {
		case reflect.String:
			if l := f.Len(); l > maxFieldString {
				return status.Errorf(
					codes.InvalidArgument,
					"exceeds size limit: %s: max=%d, size=%d",
					tv.Field(i).Name, maxFieldString, l)
			}
		case reflect.Map:
			if f.Len() == 0 {
				continue
			}
			size := 0
			for _, k := range f.MapKeys() {
				if k.Kind() == reflect.String {
					kl := k.Len()
					if kl > maxFieldString {
						return status.Errorf(
							codes.InvalidArgument,
							"exceeds size limit: %s[%s]: max=%d, size=%d",
							tv.Field(i).Name, k.String(), maxFieldString, kl)
					}
					size = size + kl
				}
				if v := f.MapIndex(k); v.Kind() == reflect.String {
					vl := v.Len()
					if vl > maxFieldString {
						return status.Errorf(
							codes.InvalidArgument,
							"exceeds size limit: %s[%s]=: max=%d, size=%d",
							tv.Field(i).Name, k.String(), maxFieldString, vl)
					}
					size = size + vl
				}
			}
			if size > maxFieldMap {
				return status.Errorf(
					codes.InvalidArgument,
					"exceeds size limit: %s: max=%d, size=%d",
					tv.Field(i).Name, maxFieldMap, size)
			}
		}
	}
	return nil
}
