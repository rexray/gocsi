package gocsi

import (
	"sync"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

// SpecValidatorOption configures the spec validator interceptor.
type SpecValidatorOption func(*specValidatorOpts)

type specValidatorOpts struct {
	sync.Mutex
	supportedVersions   []csi.Version
	requiresNodeID      bool
	requiresPubVolInfo  bool
	requiresVolAttribs  bool
	requiresCredentials map[string]struct{}
	successfulExitCodes map[string]map[codes.Code]struct{}
}

func (o *specValidatorOpts) setSuccessfulExitCode(m string, c codes.Code) {

	o.Lock()
	defer o.Unlock()
	if o.successfulExitCodes == nil {
		o.successfulExitCodes = map[string]map[codes.Code]struct{}{}
	}
	codez, ok := o.successfulExitCodes[m]
	if !ok {
		codez = map[codes.Code]struct{}{}
		o.successfulExitCodes[m] = codez
	}
	codez[c] = struct{}{}
}

func (o *specValidatorOpts) requireCredentials(m string) {
	o.Lock()
	defer o.Unlock()
	if o.requiresCredentials == nil {
		o.requiresCredentials = map[string]struct{}{}
	}
	o.requiresCredentials[m] = struct{}{}
}

// WithSupportedVersions is a SpecValidatorOption that indicates the
// list of versions supported by any CSI RPC that participates in
// version validation.
func WithSupportedVersions(versions ...csi.Version) SpecValidatorOption {
	return func(o *specValidatorOpts) {
		o.supportedVersions = versions
	}
}

// WithSuccessCreateVolumeAlreadyExists is a SpecValidatorOption that the
// eponymous request should treat the eponymous error code as successful.
func WithSuccessCreateVolumeAlreadyExists() SpecValidatorOption {
	return func(o *specValidatorOpts) {
		o.setSuccessfulExitCode(CreateVolume, codes.AlreadyExists)
	}
}

// WithSuccessDeleteVolumeNotFound is a SpecValidatorOption that the
// eponymous request should treat the eponymous error code as successful.
func WithSuccessDeleteVolumeNotFound() SpecValidatorOption {
	return func(o *specValidatorOpts) {
		o.setSuccessfulExitCode(DeleteVolume, codes.NotFound)
	}
}

// WithRequiresNodeID is a SpecValidatorOption that indicates
// ControllerPublishVolume requests and GetNodeID responses must
// contain non-empty node ID data.
func WithRequiresNodeID() SpecValidatorOption {
	return func(o *specValidatorOpts) {
		o.requiresNodeID = true
	}
}

// WithRequiresPublishVolumeInfo is a SpecValidatorOption that indicates
// ControllerPublishVolume responses and NodePublishVolume requests must
// contain non-empty publish volume info data.
func WithRequiresPublishVolumeInfo() SpecValidatorOption {
	return func(o *specValidatorOpts) {
		o.requiresPubVolInfo = true
	}
}

// WithRequiresVolumeAttributes is a SpecValidatorOption that indicates
// ControllerPublishVolume, ValidateVolumeCapabilities, and NodePublishVolume
// requests must contain non-empty volume attribute data.
func WithRequiresVolumeAttributes() SpecValidatorOption {
	return func(o *specValidatorOpts) {
		o.requiresVolAttribs = true
	}
}

// WithRequiresCreateVolumeCredentials is a SpecValidatorOption
// that indicates the eponymous requests must contain non-empty credentials
// data.
func WithRequiresCreateVolumeCredentials() SpecValidatorOption {
	return func(o *specValidatorOpts) {
		o.requireCredentials(CreateVolume)
	}
}

// WithRequiresDeleteVolumeCredentials is a SpecValidatorOption
// that indicates the eponymous requests must contain non-empty credentials
// data.
func WithRequiresDeleteVolumeCredentials() SpecValidatorOption {
	return func(o *specValidatorOpts) {
		o.requireCredentials(DeleteVolume)
	}
}

// WithRequiresControllerPublishVolumeCredentials is a SpecValidatorOption
// that indicates the eponymous requests must contain non-empty credentials
// data.
func WithRequiresControllerPublishVolumeCredentials() SpecValidatorOption {
	return func(o *specValidatorOpts) {
		o.requireCredentials(ControllerPublishVolume)
	}
}

// WithRequiresControllerUnpublishVolumeCredentials is a SpecValidatorOption
// that indicates the eponymous requests must contain non-empty credentials
// data.
func WithRequiresControllerUnpublishVolumeCredentials() SpecValidatorOption {
	return func(o *specValidatorOpts) {
		o.requireCredentials(ControllerUnpublishVolume)
	}
}

// WithRequiresNodePublishVolumeCredentials is a SpecValidatorOption
// that indicates the eponymous requests must contain non-empty credentials
// data.
func WithRequiresNodePublishVolumeCredentials() SpecValidatorOption {
	return func(o *specValidatorOpts) {
		o.requireCredentials(NodePublishVolume)
	}
}

// WithRequiresNodeUnpublishVolumeCredentials is a SpecValidatorOption
// that indicates the eponymous requests must contain non-empty credentials
// data.
func WithRequiresNodeUnpublishVolumeCredentials() SpecValidatorOption {
	return func(o *specValidatorOpts) {
		o.requireCredentials(NodeUnpublishVolume)
	}
}

type specValidator struct {
	opts specValidatorOpts
}

// NewServerSpecValidator returns a new UnaryServerInterceptor that validates
// server request and response data against the CSI specification.
func NewServerSpecValidator(
	opts ...SpecValidatorOption) grpc.UnaryServerInterceptor {

	return newSpecValidator(opts...).handleServer
}

// NewClientSpecValidator provides a UnaryClientInterceptor that validates
// client request and response data against the CSI specification.
func NewClientSpecValidator(
	opts ...SpecValidatorOption) grpc.UnaryClientInterceptor {

	return newSpecValidator(opts...).handleClient
}

func newSpecValidator(opts ...SpecValidatorOption) *specValidator {
	i := &specValidator{}
	for _, withOpts := range opts {
		withOpts(&i.opts)
	}
	return i
}

func (s *specValidator) handleServer(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	return s.handle(ctx, info.FullMethod, req, func() (interface{}, error) {
		return handler(ctx, req)
	})
}

func (s *specValidator) handleClient(
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

func (s *specValidator) handle(
	ctx context.Context,
	method string,
	req interface{},
	next func() (interface{}, error)) (interface{}, error) {

	// If the request is nil then pass control to the next handler
	// in the chain.
	if req == nil {
		return next()
	}

	// Validate the request version.
	if err := s.validateRequestVersion(ctx, req); err != nil {
		return nil, err
	}

	// Validate the request against the CSI specification.
	if err := s.validateRequest(ctx, method, req); err != nil {
		return nil, err
	}

	// Use the function passed into this one to get the response. On the
	// server-side this could possibly invoke additional interceptors or
	// the RPC. On the client side this invokes the RPC.
	rep, err := next()

	// Determine whether or not the response is nil. Otherwise it
	// will no longer be possible to perform a nil equality check on the
	// response to the interface{} rules for nil comparison.
	isNilRep := isResponseNil(method, rep)

	// Handle possible non-zero successful exit codes.
	if err := s.handleResponseError(method, err); err != nil {
		if isNilRep {
			return nil, err
		}
	}

	// If the response is nil then go ahead and return a nil value
	// directly in order to fulfill Go's rules about nil values and
	// interface{} types. For more information please see the links
	// in the previous comment.
	if isNilRep {
		return nil, nil
	}

	// Validate the response against the CSI specification.
	if err := s.validateResponse(ctx, method, rep); err != nil {
		return rep, err
	}

	return rep, err
}

func (s *specValidator) handleResponseError(method string, err error) error {

	// If the returned error does not contain a gRPC error code then
	// return early from this function.
	stat, ok := status.FromError(err)
	if !ok {
		return err
	}

	// Error code OK always equals success, so clear the error.
	if stat.Code() == codes.OK {
		return nil
	}

	// Check to see if the current method is configured to treat
	// any non-zero exit codes as successful. If so, and the current
	// exit code matches any of them, then clear the error.
	for exitCode := range s.opts.successfulExitCodes[method] {
		if stat.Code() == exitCode {
			log.WithFields(log.Fields{
				"code": stat.Code(),
				"msg":  stat.Message(),
			}).Debug("dropping error")
			return nil
		}
	}

	return err
}

type specValidatorHasVolumeID interface {
	GetVolumeId() string
}
type specValidatorHasNodeID interface {
	GetNodeId() string
}
type specValidatorHasUserCredentials interface {
	GetUserCredentials() map[string]string
}
type specValidatorHasVolumeAttributes interface {
	GetVolumeAttributes() map[string]string
}
type specValidatorHasVersion interface {
	GetVersion() *csi.Version
}

func (s *specValidator) validateRequest(
	ctx context.Context,
	method string,
	req interface{}) error {

	if req == nil {
		return nil
	}

	// Check to see if the request has a volume ID and if it is set.
	// If the volume ID is not set then return an error.
	if treq, ok := req.(specValidatorHasVolumeID); ok {
		if treq.GetVolumeId() == "" {
			return ErrVolumeIDRequired
		}
	}

	// Check to see if the request has a node ID and if it is set.
	// If the node ID is not set then return an error.
	if treq, ok := req.(specValidatorHasNodeID); ok {
		if treq.GetNodeId() == "" {
			return ErrNodeIDRequired
		}
	}

	// Check to see if the request has credentials and if they're required.
	// If the credentials are required but no credentials are specified then
	// return an error.
	if _, ok := s.opts.requiresCredentials[method]; ok {
		if treq, ok := req.(specValidatorHasUserCredentials); ok {
			if len(treq.GetUserCredentials()) == 0 {
				return ErrUserCredentialsRequired
			}
		}
	}

	// Check to see if the request has volume attributes and if they're
	// required. If the volume attributes are required by no attributes are
	// specified then return an error.
	if s.opts.requiresVolAttribs {
		if treq, ok := req.(specValidatorHasVolumeAttributes); ok {
			if len(treq.GetVolumeAttributes()) == 0 {
				return ErrVolumeAttributesRequired
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

func (s *specValidator) validateResponse(
	ctx context.Context,
	method string,
	rep interface{}) error {

	if rep == nil {
		return nil
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

func (s *specValidator) validateRequestVersion(
	ctx context.Context,
	req interface{}) error {

	// Check to see if the request version should be validated.
	if len(s.opts.supportedVersions) == 0 {
		return nil
	}

	treq, ok := req.(specValidatorHasVersion)
	if !ok {
		return nil
	}

	var (
		supported      bool
		requestVersion = treq.GetVersion()
	)

	if requestVersion == nil {
		return status.Error(
			codes.InvalidArgument, "invalid request version: nil")
	}

	for _, supportedVersion := range s.opts.supportedVersions {
		if CompareVersions(requestVersion, &supportedVersion) == 0 {
			supported = true
			break
		}
	}

	if !supported {
		return status.Errorf(
			codes.InvalidArgument,
			"invalid request version: %s",
			SprintfVersion(*requestVersion))
	}

	return nil
}

func (s *specValidator) validateCreateVolumeRequest(
	ctx context.Context,
	req csi.CreateVolumeRequest) error {

	if req.Name == "" {
		return ErrVolumeNameRequired
	}

	return validateVolumeCapabilitiesArg(req.VolumeCapabilities, true)
}

// func (s *specValidator) validateDeleteVolumeRequest(
// 	ctx context.Context,
// 	req csi.DeleteVolumeRequest) error {
//
// 	return nil
// }

func (s *specValidator) validateControllerPublishVolumeRequest(
	ctx context.Context,
	req csi.ControllerPublishVolumeRequest) error {

	return validateVolumeCapabilityArg(req.VolumeCapability, true)
}

func (s *specValidator) validateValidateVolumeCapabilitiesRequest(
	ctx context.Context,
	req csi.ValidateVolumeCapabilitiesRequest) error {

	return validateVolumeCapabilitiesArg(req.VolumeCapabilities, true)
}

func (s *specValidator) validateGetCapacityRequest(
	ctx context.Context,
	req csi.GetCapacityRequest) error {

	return validateVolumeCapabilitiesArg(req.VolumeCapabilities, false)
}

func (s *specValidator) validateNodePublishVolumeRequest(
	ctx context.Context,
	req csi.NodePublishVolumeRequest) error {

	if req.TargetPath == "" {
		return ErrTargetPathRequired
	}

	if s.opts.requiresPubVolInfo && len(req.PublishVolumeInfo) == 0 {
		return ErrPublishVolumeInfoRequired
	}

	return validateVolumeCapabilityArg(req.VolumeCapability, true)
}

func (s *specValidator) validateNodeUnpublishVolumeRequest(
	ctx context.Context,
	req csi.NodeUnpublishVolumeRequest) error {

	if req.TargetPath == "" {
		return ErrTargetPathRequired
	}

	return nil
}

func (s *specValidator) validateCreateVolumeResponse(
	ctx context.Context,
	rep csi.CreateVolumeResponse) error {

	if rep.VolumeInfo == nil {
		return ErrNilVolumeInfo
	}

	if rep.VolumeInfo.Id == "" {
		return ErrEmptyVolumeID
	}

	if s.opts.requiresVolAttribs && len(rep.VolumeInfo.Attributes) == 0 {
		return ErrNonNilEmptyAttribs
	}

	return nil
}

func (s *specValidator) validateControllerPublishVolumeResponse(
	ctx context.Context,
	rep csi.ControllerPublishVolumeResponse) error {

	if s.opts.requiresPubVolInfo && len(rep.PublishVolumeInfo) == 0 {
		return ErrEmptyPublishVolumeInfo
	}
	return nil
}

func (s *specValidator) validateListVolumesResponse(
	ctx context.Context,
	rep csi.ListVolumesResponse) error {

	for i, e := range rep.Entries {
		volInfo := e.VolumeInfo
		if volInfo == nil {
			return status.Errorf(
				codes.Internal,
				"volumeInfo is nil: index=%d", i)
		}
		if volInfo.Id == "" {
			return status.Errorf(
				codes.Internal,
				"volumeInfo.Id is empty: index=%d", i)
		}
		if volInfo.Attributes != nil && len(volInfo.Attributes) == 0 {
			return status.Errorf(
				codes.Internal,
				"volumeInfo.Attributes is not nil & empty: index=%d", i)
		}
	}

	return nil
}

func (s *specValidator) validateControllerGetCapabilitiesResponse(
	ctx context.Context,
	rep csi.ControllerGetCapabilitiesResponse) error {

	if rep.Capabilities != nil && len(rep.Capabilities) == 0 {
		return ErrNonNilControllerCapabilities
	}
	return nil
}

func (s *specValidator) validateGetSupportedVersionsResponse(
	ctx context.Context,
	rep csi.GetSupportedVersionsResponse) error {

	if len(rep.SupportedVersions) == 0 {
		return ErrEmptySupportedVersions
	}
	return nil
}

func (s *specValidator) validateGetPluginInfoResponse(
	ctx context.Context,
	rep csi.GetPluginInfoResponse) error {

	if rep.Name == "" {
		return ErrEmptyPluginName
	}
	if rep.VendorVersion == "" {
		return ErrEmptyVendorVersion
	}
	if rep.Manifest != nil && len(rep.Manifest) == 0 {
		return ErrNonNilEmptyPluginManifest
	}
	return nil
}

func (s *specValidator) validateGetNodeIDResponse(
	ctx context.Context,
	rep csi.GetNodeIDResponse) error {

	if rep.NodeId == "" {
		return ErrEmptyNodeID
	}
	return nil
}

func (s *specValidator) validateNodeGetCapabilitiesResponse(
	ctx context.Context,
	rep csi.NodeGetCapabilitiesResponse) error {

	if rep.Capabilities != nil && len(rep.Capabilities) == 0 {
		return ErrNonNilNodeCapabilities
	}
	return nil
}

func validateVolumeCapabilityArg(
	volCap *csi.VolumeCapability,
	required bool) error {

	if required && volCap == nil {
		return ErrVolumeCapabilityRequired
	}

	if volCap.AccessMode == nil {
		return ErrAccessModeRequired
	}

	atype := volCap.GetAccessType()
	if atype == nil {
		return ErrAccessTypeRequired
	}

	switch tatype := atype.(type) {
	case *csi.VolumeCapability_Block:
		if tatype.Block == nil {
			return ErrBlockTypeRequired
		}
	case *csi.VolumeCapability_Mount:
		if tatype.Mount == nil {
			return ErrMountTypeRequired
		}
	default:
		return status.Errorf(
			codes.InvalidArgument,
			"invalid access type: %T", atype)
	}

	return nil
}

func validateVolumeCapabilitiesArg(
	volCaps []*csi.VolumeCapability,
	required bool) error {

	if len(volCaps) == 0 {
		if required {
			return ErrVolumeCapabilitiesRequired
		}
		return nil
	}

	for i, cap := range volCaps {
		if cap.AccessMode == nil {
			return status.Errorf(
				codes.InvalidArgument,
				"access mode required: index %d", i)
		}
		atype := cap.GetAccessType()
		if atype == nil {
			return status.Errorf(
				codes.InvalidArgument,
				"access type: index %d required", i)
		}
		switch tatype := atype.(type) {
		case *csi.VolumeCapability_Block:
			if tatype.Block == nil {
				return status.Errorf(
					codes.InvalidArgument,
					"block type: index %d required", i)
			}
		case *csi.VolumeCapability_Mount:
			if tatype.Mount == nil {
				return status.Errorf(
					codes.InvalidArgument,
					"mount type: index %d required", i)
			}
		default:
			return status.Errorf(
				codes.InvalidArgument,
				"invalid access type: index %d, type=%T", i, atype)
		}
	}

	return nil
}
