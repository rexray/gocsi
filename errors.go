package gocsi

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrVolumeNotFound returns an error indicating a volume with the
// specified ID cannot be found.
func ErrVolumeNotFound(id string) error {
	return status.Errorf(codes.NotFound, "volume not found: %s", id)
}

// ErrMissingCSIEndpoint occurs when the value for the environment
// variable CSI_ENDPOINT is not set.
var ErrMissingCSIEndpoint = errors.New("missing CSI_ENDPOINT")

// ErrOpPending occurs when an RPC is made against a resource that
// is involved in a concurrent operation.
var ErrOpPending = status.Error(
	codes.FailedPrecondition, "op pending")

// ErrVolumeIDRequired occurs when an RPC is made with an empty
// volume ID argument.
var ErrVolumeIDRequired = status.Error(
	codes.InvalidArgument, "volume ID required")

// ErrNodeIDRequired occurs when an RPC is made with an empty
// node ID argument.
var ErrNodeIDRequired = status.Error(
	codes.InvalidArgument, "node ID required")

// ErrVolumeNameRequired occurs when an RPC is made with an empty
// volume name argument.
var ErrVolumeNameRequired = status.Error(
	codes.InvalidArgument, "volume name required")

// ErrVolumeCapabilityRequired occurs when an RPC is made with
// a missing volume capability argument.
var ErrVolumeCapabilityRequired = status.Error(
	codes.InvalidArgument, "volume capability required")

// ErrAccessModeRequired occurs when an RPC is made with
// a missing acess mode argument.
var ErrAccessModeRequired = status.Error(
	codes.InvalidArgument, "acess mode required")

// ErrAccessTypeRequired occurs when an RPC is made with
// a missing acess type argument.
var ErrAccessTypeRequired = status.Error(
	codes.InvalidArgument, "acess type required")

// ErrBlockTypeRequired occurs when an RPC is made with
// a missing access type block value.
var ErrBlockTypeRequired = status.Error(
	codes.InvalidArgument, "block type required")

// ErrMountTypeRequired occurs when an RPC is made with
// a missing access type mount value.
var ErrMountTypeRequired = status.Error(
	codes.InvalidArgument, "mount type required")

// ErrVolumeCapabilitiesRequired occurs when an RPC is made with
// an empty volume capabilties argument.
var ErrVolumeCapabilitiesRequired = status.Error(
	codes.InvalidArgument, "volume capabilities required")

// ErrUserCredentialsRequired occurs when an RPC is made with
// an empty user credentials argument.
var ErrUserCredentialsRequired = status.Error(
	codes.InvalidArgument, "user credentials required")

// ErrVolumeAttributesRequired occurs when an RPC is made with
// an empty volume attributes argument.
var ErrVolumeAttributesRequired = status.Error(
	codes.InvalidArgument, "volume attributes required")

// ErrPublishVolumeInfoRequired occurs when an RPC is made with
// an empty publish volume info argument.
var ErrPublishVolumeInfoRequired = status.Error(
	codes.InvalidArgument, "publish volume info required")

// ErrTargetPathRequired occurs when an RPC is made with an empty
// target path argument.
var ErrTargetPathRequired = status.Error(
	codes.InvalidArgument, "target path required")

// ErrInvalidTargetPath occurs when an RPC is made with
// an invalid targetPath argument.
var ErrInvalidTargetPath = errors.New("invalid targetPath")

// ErrNilVolumeInfo occurs when an RPC returns a nil VolumeInfo.
var ErrNilVolumeInfo = status.Error(
	codes.Internal, "nil volumeInfo")

// ErrEmptyVolumeID occurs when an RPC returns a VolumeInfo with
// an zero-length Id field.
var ErrEmptyVolumeID = status.Error(
	codes.Internal, "empty volumeInfo.Id")

// ErrNonNilEmptyAttribs occurs when an RPC returns a VolumeInfo
// with a non-nil Attributes field that has no elements.
var ErrNonNilEmptyAttribs = status.Error(
	codes.Internal, "non-nil, empty volumeInfo.Attributes")

// ErrEmptyPublishVolumeInfo occurs when an RPC returns
// empty publish volume info.
var ErrEmptyPublishVolumeInfo = status.Error(
	codes.Internal, "empty publish volume info")

// ErrEmptyNodeID occurs when an RPC returns an empty NodeID.
var ErrEmptyNodeID = status.Error(codes.Internal, "empty node ID")

// ErrEmptySupportedVersions occurs when an RPC returns a zero-length
// supported versions list.
var ErrEmptySupportedVersions = status.Error(
	codes.Internal, "empty supported versions")

// ErrEmptyPluginName occurs when GetPluginInfo returns an empty
// plug-in name.
var ErrEmptyPluginName = status.Error(
	codes.Internal, "empty plug-in name")

// ErrEmptyVendorVersion occurs when GetPluginInfo returns an empty
// vendor version.
var ErrEmptyVendorVersion = status.Error(
	codes.Internal, "empty vendor version")

// ErrNonNilEmptyPluginManifest occurs when GetPluginInfo returns a non-nil,
// empty manifest.
var ErrNonNilEmptyPluginManifest = status.Error(
	codes.Internal, "non-nil, empty plug-in manifest")

// ErrNonNilControllerCapabilities occurs when NodeGetCapabilities returns
// a non-nil, empty list.
var ErrNonNilControllerCapabilities = status.Error(
	codes.Internal, "non-nil, empty controller capabilities")

// ErrNonNilNodeCapabilities occurs when NodeGetCapabilities returns
// a non-nil, empty list.
var ErrNonNilNodeCapabilities = status.Error(
	codes.Internal, "non-nil, empty node capabilities")

// ErrInvalidProvider is returned from NewService if the
// specified provider name is unknown.
var ErrInvalidProvider = errors.New("invalid service provider")

// IsSuccess returns nil if the provided error is an RPC error with an error
// code that is OK (0) or matches one of the additional, provided successful
// error codes. Otherwise the original error is returned.
func IsSuccess(err error, successCodes ...codes.Code) error {

	// Shortcut the process by first checking to see if the error is nil.
	if err == nil {
		return nil
	}

	// Check to see if the provided error is an RPC error.
	stat, ok := status.FromError(err)
	if !ok {
		return err
	}

	if stat.Code() == codes.OK {
		return nil
	}
	for _, c := range successCodes {
		if stat.Code() == c {
			return nil
		}
	}

	return err
}

// IsSuccessfulResponse uses IsSuccess to determine if the response for
// a specific CSI method is successful. If successful a nil value is
// returned; otherwise the original error is returned.
func IsSuccessfulResponse(method string, err error) error {
	switch method {
	case CreateVolume:
		return IsSuccess(err, codes.AlreadyExists)
	case DeleteVolume:
		return IsSuccess(err, codes.NotFound)
	}
	return err
}

func notFound(e error) error {
	if s, k := status.FromError(e); k && s.Code() == codes.NotFound {
		return nil
	}
	return e
}

func alreadyExists(e error) error {
	if s, k := status.FromError(e); k && s.Code() == codes.AlreadyExists {
		return nil
	}
	return e
}
