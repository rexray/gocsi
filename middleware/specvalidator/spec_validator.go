package specvalidator

import (
	"reflect"
	"regexp"
	"strconv"
	"sync"

	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"

	csienv "github.com/rexray/gocsi/env"
	"github.com/rexray/gocsi/utils"
)

// Middleware injects a unique ID into outgoing requests and reads the
// ID from incoming requests.
type Middleware struct {
	sync.Once
	RequestValidation                       bool
	ResponseValidation                      bool
	RequireStagingTargetPath                bool
	RequireNodeID                           bool
	RequirePublishInfo                      bool
	RequireVolumeAttributes                 bool
	RequireControllerCreateVolumeSecrets    bool
	RequireControllerDeleteVolumeSecrets    bool
	RequireControllerPublishVolumeSecrets   bool
	RequireControllerUnpublishVolumeSecrets bool
	RequireNodeStageVolumeSecrets           bool
	RequireNodePublishVolumeSecrets         bool
}

// Init is available to explicitly initialize the middleware.
func (s *Middleware) Init(ctx context.Context) (err error) {
	return s.initOnce(ctx)
}

func (s *Middleware) initOnce(ctx context.Context) (err error) {
	s.Once.Do(func() {
		err = s.init(ctx)
	})
	return
}

func (s *Middleware) init(ctx context.Context) error {

	setBool := func(addr *bool, key string) {
		v, ok := csienv.LookupEnv(ctx, key)
		if !ok {
			return
		}
		b, err := strconv.ParseBool(v)
		if err != nil {
			return
		}
		*addr = b
		log.WithField(key, b).Debug("middleware: spec validation")
		if b {
			s.RequestValidation = true
		}
	}
	setBool(
		&s.RequireStagingTargetPath,
		"X_CSI_REQUIRE_STAGING_TARGET_PATH")
	setBool(
		&s.RequireNodeID,
		"X_CSI_REQUIRE_NODE_ID")
	setBool(
		&s.RequirePublishInfo,
		"X_CSI_REQUIRE_PUB_INFO")
	setBool(
		&s.RequireVolumeAttributes,
		"X_CSI_REQUIRE_VOL_ATTRIBS")
	setBool(
		&s.RequireNodeID,
		"X_CSI_REQUIRE_NODE_ID")

	if v, ok := csienv.LookupEnv(ctx, "X_CSI_REQUIRE_SECRETS"); ok {
		if b, _ := strconv.ParseBool(v); b {
			s.RequireControllerCreateVolumeSecrets = true
			s.RequireControllerDeleteVolumeSecrets = true
			s.RequireControllerPublishVolumeSecrets = true
			s.RequireControllerUnpublishVolumeSecrets = true
			s.RequireNodeStageVolumeSecrets = true
			s.RequireNodePublishVolumeSecrets = true
		}
	}
	setBool(
		&s.RequireControllerCreateVolumeSecrets,
		"X_CSI_REQUIRE_SECRETS_CTRLR_CREATE_VOL")
	setBool(
		&s.RequireControllerDeleteVolumeSecrets,
		"X_CSI_REQUIRE_SECRETS_CTRLR_DELETE_VOL")
	setBool(
		&s.RequireControllerPublishVolumeSecrets,
		"X_CSI_REQUIRE_SECRETS_CTRLR_PUB_VOL")
	setBool(
		&s.RequireControllerUnpublishVolumeSecrets,
		"X_CSI_REQUIRE_SECRETS_CTRLR_UNPUB_VOL")
	setBool(
		&s.RequireNodeStageVolumeSecrets,
		"X_CSI_REQUIRE_SECRETS_NODE_STAGE_VOL")
	setBool(
		&s.RequireNodePublishVolumeSecrets,
		"X_CSI_REQUIRE_SECRETS_NODE_PUB_VOL")

	if v, ok := csienv.LookupEnv(ctx, "X_CSI_SPEC_VALIDATION"); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			s.RequestValidation = b
			s.ResponseValidation = b
		}
	}
	// Even if request validation is implicitly enabled, it can be
	// explicitly disabled.
	setBool(
		&s.RequestValidation,
		"X_CSI_SPEC_REQ_VALIDATION")
	setBool(
		&s.ResponseValidation,
		"X_CSI_SPEC_REP_VALIDATION")

	if s.RequestValidation || s.ResponseValidation {
		log.WithFields(map[string]interface{}{
			"RequestValidation":                       s.RequestValidation,
			"ResponseValidation":                      s.ResponseValidation,
			"RequireStagingTargetPath":                s.RequireStagingTargetPath,
			"RequireNodeID":                           s.RequireNodeID,
			"RequirePublishInfo":                      s.RequirePublishInfo,
			"RequireVolumeAttributes":                 s.RequireVolumeAttributes,
			"RequireControllerCreateVolumeSecrets":    s.RequireControllerCreateVolumeSecrets,
			"RequireControllerDeleteVolumeSecrets":    s.RequireControllerDeleteVolumeSecrets,
			"RequireControllerPublishVolumeSecrets":   s.RequireControllerPublishVolumeSecrets,
			"RequireControllerUnpublishVolumeSecrets": s.RequireControllerUnpublishVolumeSecrets,
			"RequireNodeStageVolumeSecrets":           s.RequireNodeStageVolumeSecrets,
			"RequireNodePublishVolumeSecrets":         s.RequireNodePublishVolumeSecrets,
		}).Info("middleware: spec validation")
	}

	return nil
}

// HandleServer is a UnaryServerInterceptor that validates
// server request and response data against the CSI specification.
func (s *Middleware) HandleServer(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	return s.handle(ctx, info.FullMethod, req, func() (interface{}, error) {
		return handler(ctx, req)
	})
}

// HandleClient is a UnaryClientInterceptor that validates
// client request and response data against the CSI specification.
func (s *Middleware) HandleClient(
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

func (s *Middleware) handle(
	ctx context.Context,
	method string,
	req interface{},
	next func() (interface{}, error)) (interface{}, error) {

	if err := s.initOnce(ctx); err != nil {
		return nil, err
	}

	// If the request is nil then pass control to the next handler
	// in the chain.
	if req == nil {
		return next()
	}

	if s.RequestValidation {
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

	if s.ResponseValidation {
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
	NodeGetId() string
}
type interceptorHasUserCredentials interface {
	GetUserCredentials() map[string]string
}
type interceptorHasVolumeAttributes interface {
	GetVolumeAttributes() map[string]string
}

func (s *Middleware) validateRequest(
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
	if s.RequireNodeID {
		if treq, ok := req.(interceptorHasNodeID); ok {
			if treq.NodeGetId() == "" {
				return status.Error(
					codes.InvalidArgument, "required: NodeID")
			}
		}
	}

	// Check to see if the request has volume attributes and if they're
	// required. If the volume attributes are required by no attributes are
	// specified then return an error.
	if s.RequireVolumeAttributes {
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
	case *csi.DeleteVolumeRequest:
		return s.validateDeleteVolumeRequest(ctx, *tobj)
	case *csi.ControllerPublishVolumeRequest:
		return s.validateControllerPublishVolumeRequest(ctx, *tobj)
	case *csi.ControllerUnpublishVolumeRequest:
		return s.validateControllerUnpublishVolumeRequest(ctx, *tobj)
	case *csi.ValidateVolumeCapabilitiesRequest:
		return s.validateValidateVolumeCapabilitiesRequest(ctx, *tobj)
	case *csi.GetCapacityRequest:
		return s.validateGetCapacityRequest(ctx, *tobj)
		//
		// Node Service
		//
	case *csi.NodeStageVolumeRequest:
		return s.validateNodeStageVolumeRequest(ctx, *tobj)
	case *csi.NodePublishVolumeRequest:
		return s.validateNodePublishVolumeRequest(ctx, *tobj)
	case *csi.NodeUnpublishVolumeRequest:
		return s.validateNodeUnpublishVolumeRequest(ctx, *tobj)
	}

	return nil
}

func (s *Middleware) validateResponse(
	ctx context.Context,
	method string,
	rep interface{}) error {

	if utils.IsNilResponse(rep) {
		return status.Error(codes.Internal, "nil response")
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
	case *csi.GetPluginInfoResponse:
		return s.validateGetPluginInfoResponse(ctx, *tobj)
	//
	// Node Service
	//
	case *csi.NodeGetIdResponse:
		return s.validateNodeGetIDResponse(ctx, *tobj)
	case *csi.NodeGetCapabilitiesResponse:
		return s.validateNodeGetCapabilitiesResponse(ctx, *tobj)
	}

	return nil
}

func (s *Middleware) validateCreateVolumeRequest(
	ctx context.Context,
	req csi.CreateVolumeRequest) error {

	if req.Name == "" {
		return status.Error(
			codes.InvalidArgument, "required: Name")
	}
	if s.RequireControllerCreateVolumeSecrets {
		if len(req.ControllerCreateSecrets) == 0 {
			return status.Error(
				codes.InvalidArgument, "required: ControllerCreateSecrets")
		}
	}

	return validateVolumeCapabilitiesArg(req.VolumeCapabilities, true)
}

func (s *Middleware) validateDeleteVolumeRequest(
	ctx context.Context,
	req csi.DeleteVolumeRequest) error {

	if s.RequireControllerDeleteVolumeSecrets {
		if len(req.ControllerDeleteSecrets) == 0 {
			return status.Error(
				codes.InvalidArgument, "required: ControllerDeleteSecrets")
		}
	}

	return nil
}

func (s *Middleware) validateControllerPublishVolumeRequest(
	ctx context.Context,
	req csi.ControllerPublishVolumeRequest) error {

	if s.RequireControllerPublishVolumeSecrets {
		if len(req.ControllerPublishSecrets) == 0 {
			return status.Error(
				codes.InvalidArgument, "required: ControllerPublishSecrets")
		}
	}

	return validateVolumeCapabilityArg(req.VolumeCapability, true)
}

func (s *Middleware) validateControllerUnpublishVolumeRequest(
	ctx context.Context,
	req csi.ControllerUnpublishVolumeRequest) error {

	if s.RequireControllerUnpublishVolumeSecrets {
		if len(req.ControllerUnpublishSecrets) == 0 {
			return status.Error(
				codes.InvalidArgument, "required: ControllerUnpublishSecrets")
		}
	}

	return nil
}

func (s *Middleware) validateValidateVolumeCapabilitiesRequest(
	ctx context.Context,
	req csi.ValidateVolumeCapabilitiesRequest) error {

	return validateVolumeCapabilitiesArg(req.VolumeCapabilities, true)
}

func (s *Middleware) validateGetCapacityRequest(
	ctx context.Context,
	req csi.GetCapacityRequest) error {

	return validateVolumeCapabilitiesArg(req.VolumeCapabilities, false)
}

func (s *Middleware) validateNodeStageVolumeRequest(
	ctx context.Context,
	req csi.NodeStageVolumeRequest) error {

	if req.StagingTargetPath == "" {
		return status.Error(
			codes.InvalidArgument, "required: StagingTargetPath")
	}

	if s.RequirePublishInfo && len(req.PublishInfo) == 0 {
		return status.Error(
			codes.InvalidArgument, "required: PublishInfo")
	}

	if s.RequireNodeStageVolumeSecrets {
		if len(req.NodeStageSecrets) == 0 {
			return status.Error(
				codes.InvalidArgument, "required: NodeStageSecrets")
		}
	}

	return validateVolumeCapabilityArg(req.VolumeCapability, true)
}

func (s *Middleware) validateNodePublishVolumeRequest(
	ctx context.Context,
	req csi.NodePublishVolumeRequest) error {

	if s.RequireStagingTargetPath && req.StagingTargetPath == "" {
		return status.Error(
			codes.InvalidArgument, "required: StagingTargetPath")
	}

	if req.TargetPath == "" {
		return status.Error(
			codes.InvalidArgument, "required: TargetPath")
	}

	if s.RequirePublishInfo && len(req.PublishInfo) == 0 {
		return status.Error(
			codes.InvalidArgument, "required: PublishInfo")
	}

	if s.RequireNodePublishVolumeSecrets {
		if len(req.NodePublishSecrets) == 0 {
			return status.Error(
				codes.InvalidArgument, "required: NodePublishSecrets")
		}
	}

	return validateVolumeCapabilityArg(req.VolumeCapability, true)
}

func (s *Middleware) validateNodeUnpublishVolumeRequest(
	ctx context.Context,
	req csi.NodeUnpublishVolumeRequest) error {

	if req.TargetPath == "" {
		return status.Error(
			codes.InvalidArgument, "required: TargetPath")
	}

	return nil
}

func (s *Middleware) validateCreateVolumeResponse(
	ctx context.Context,
	rep csi.CreateVolumeResponse) error {

	if rep.Volume == nil {
		return status.Error(codes.Internal, "nil: Volume")
	}

	if rep.Volume.Id == "" {
		return status.Error(codes.Internal, "empty: Volume.Id")
	}

	if s.RequireVolumeAttributes && len(rep.Volume.Attributes) == 0 {
		return status.Error(
			codes.Internal, "non-nil, empty: Volume.Attributes")
	}

	return nil
}

func (s *Middleware) validateControllerPublishVolumeResponse(
	ctx context.Context,
	rep csi.ControllerPublishVolumeResponse) error {

	if s.RequirePublishInfo && len(rep.PublishInfo) == 0 {
		return status.Error(codes.Internal, "empty: PublishInfo")
	}
	return nil
}

func (s *Middleware) validateListVolumesResponse(
	ctx context.Context,
	rep csi.ListVolumesResponse) error {

	for i, e := range rep.Entries {
		vol := e.Volume
		if vol == nil {
			return status.Errorf(
				codes.Internal,
				"nil: Entries[%d].Volume", i)
		}
		if vol.Id == "" {
			return status.Errorf(
				codes.Internal,
				"empty: Entries[%d].Volume.Id", i)
		}
		if vol.Attributes != nil && len(vol.Attributes) == 0 {
			return status.Errorf(
				codes.Internal,
				"non-nil, empty: Entries[%d].Volume.Attributes", i)
		}
	}

	return nil
}

func (s *Middleware) validateControllerGetCapabilitiesResponse(
	ctx context.Context,
	rep csi.ControllerGetCapabilitiesResponse) error {

	if rep.Capabilities != nil && len(rep.Capabilities) == 0 {
		return status.Error(codes.Internal, "non-nil, empty: Capabilities")
	}
	return nil
}

const (
	pluginNameMax           = 63
	pluginNamePatt          = `^[\w\d]+\.[\w\d\.\-_]*[\w\d]$`
	pluginVendorVersionPatt = `^v?(\d+\.){2}(\d+)(-.+)?$`
)

func (s *Middleware) validateGetPluginInfoResponse(
	ctx context.Context,
	rep csi.GetPluginInfoResponse) error {

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

func (s *Middleware) validateNodeGetIDResponse(
	ctx context.Context,
	rep csi.NodeGetIdResponse) error {

	if s.RequireNodeID && rep.NodeId == "" {
		return status.Error(codes.Internal, "empty: NodeID")
	}
	return nil
}

func (s *Middleware) validateNodeGetCapabilitiesResponse(
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

// Usage returns the middleware's usage string.
func (s *Middleware) Usage() string {
	return usage
}

const usage = `SPEC VALIDATION
    X_CSI_SPEC_VALIDATION
        Setting X_CSI_SPEC_VALIDATION=true is the same as:
            X_CSI_SPEC_REQ_VALIDATION=true
            X_CSI_SPEC_REP_VALIDATION=true

    X_CSI_SPEC_REQ_VALIDATION
        A flag that enables the validation of CSI request messages.

    X_CSI_SPEC_REP_VALIDATION
        A flag that enables the validation of CSI response messages.
        Invalid responses are marshalled into a gRPC error with a code
        of "Internal."

    X_CSI_REQUIRE_STAGING_TARGET_PATH
        A flag that enables treating the following fields as required:
            * NodePublishVolumeRequest.StagingTargetPath

        Enabling this option sets X_CSI_SPEC_REQ_VALIDATION=true.

    X_CSI_REQUIRE_NODE_ID
        A flag that enables treating the following fields as required:
            * ControllerPublishVolumeRequest.NodeId
            * NodeGetIdResponse.NodeId

        Enabling this option sets X_CSI_SPEC_REQ_VALIDATION=true.

    X_CSI_REQUIRE_PUB_INFO
        A flag that enables treating the following fields as required:
            * ControllerPublishVolumeResponse.PublishInfo
            * NodeStageVolumeRequest.PublishInfo
            * NodePublishVolumeRequest.PublishInfo

        Enabling this option sets X_CSI_SPEC_REQ_VALIDATION=true.

    X_CSI_REQUIRE_VOL_ATTRIBS
        A flag that enables treating the following fields as required:
            * ControllerPublishVolumeRequest.VolumeAttributes
            * ValidateVolumeCapabilitiesRequest.VolumeAttributes
            * NodePublishVolumeRequest.VolumeAttributes

        Enabling this option sets X_CSI_SPEC_REQ_VALIDATION=true.

    X_CSI_REQUIRE_SECRETS
        Setting X_CSI_REQUIRE_SECRETS=true is the same as:
            X_CSI_REQUIRE_SECRETS_CTRLR_CREATE_VOL=true
            X_CSI_REQUIRE_SECRETS_CTRLR_DELETE_VOL=true
            X_CSI_REQUIRE_SECRETS_CTRLR_PUB_VOL=true
            X_CSI_REQUIRE_SECRETS_CTRLR_UNPUB_VOL=true
            X_CSI_REQUIRE_SECRETS_NODE_STAGE_VOL=true
            X_CSI_REQUIRE_SECRETS_NODE_PUB_VOL=true

        Enabling this option sets X_CSI_SPEC_REQ_VALIDATION=true.

    X_CSI_REQUIRE_SECRETS_CTRLR_CREATE_VOL
        A flag that enables treating the following fields as required:
            * CreateVolumeRequest.UserCredentials

        Enabling this option sets X_CSI_SPEC_REQ_VALIDATION=true.

    X_CSI_REQUIRE_SECRETS_CTRLR_DELETE_VOL
        A flag that enables treating the following fields as required:
            * DeleteVolumeRequest.UserCredentials

        Enabling this option sets X_CSI_SPEC_REQ_VALIDATION=true.

    X_CSI_REQUIRE_SECRETS_CTRLR_PUB_VOL
        A flag that enables treating the following fields as required:
            * ControllerPublishVolumeRequest.UserCredentials

        Enabling this option sets X_CSI_SPEC_REQ_VALIDATION=true.

    X_CSI_REQUIRE_SECRETS_CTRLR_UNPUB_VOL
        A flag that enables treating the following fields as required:
            * ControllerUnpublishVolumeRequest.UserCredentials

        Enabling this option sets X_CSI_SPEC_REQ_VALIDATION=true.

    X_CSI_REQUIRE_SECRETS_NODE_STAGE_VOL
        A flag that enables treating the following fields as required:
            * NodeStageVolumeRequest.UserCredentials

    X_CSI_REQUIRE_SECRETS_NODE_PUB_VOL
        A flag that enables treating the following fields as required:
            * NodePublishVolumeRequest.UserCredentials

        Enabling this option sets X_CSI_SPEC_REQ_VALIDATION=true.`
