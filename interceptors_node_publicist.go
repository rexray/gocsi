package gocsi

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/container-storage-interface/spec/lib/go/csi"
	log "github.com/sirupsen/logrus"
	"github.com/thecodeteam/gofsutil"
	xctx "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FormatAndMountFunc is returned by
// NodeVolumePublicationProvider.FormatAndMount.
type FormatAndMountFunc func(
	ctx context.Context,
	devicePath, targetPath, fsType string,
	mntFlags []string,
	userCreds map[string]string) error

// NodeVolumePublicationProvider is the interface that works with a
// server-side, gRPC interceptor to handle publishing and unpublishing
// volumes on a Node host for the SP.
type NodeVolumePublicationProvider interface {

	// GetDevicePath returns the path to a volume's device on a node host.
	GetDevicePath(
		ctx context.Context,
		volID string,
		pubVolInfo, userCreds map[string]string) (string, error)

	// GetPrivateMountTargetName returns the name of a volume's private mount
	// target. If this function returns an empty string then the
	// name of the volume's private mount target will be the MD5 checksum
	// of the volume ID.
	GetPrivateMountTargetName(
		ctx context.Context,
		volID string,
		userCreds map[string]string) (string, error)

	// FormatAndMount returns a function used to format and mount
	// volumes with a Mount capability. If the return value is nil then
	// the interceptor performs the operation.
	FormatAndMount() FormatAndMountFunc
}

type nodeVolumePublicist struct {
	p               NodeVolumePublicationProvider
	privateMountDir string
	validateDevice  bool
}

// NewNodeVolumePublicist returns a new UnaryServerInterceptor that handles
// volume publication on node hosts.
func NewNodeVolumePublicist(
	p NodeVolumePublicationProvider,
	privateMountDir string) grpc.UnaryServerInterceptor {

	return (&nodeVolumePublicist{p: p, privateMountDir: privateMountDir}).handle
}

func (i *nodeVolumePublicist) handle(
	ctx xctx.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	switch treq := req.(type) {
	case *csi.NodePublishVolumeRequest:
		if err := i.publish(ctx, treq); err != nil {
			return nil, err
		}
		return &csi.NodePublishVolumeResponse{}, nil
	case *csi.NodeUnpublishVolumeRequest:
		if err := i.unpublish(ctx, treq); err != nil {
			return nil, err
		}
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}
	return handler(ctx, req)
}

func (i *nodeVolumePublicist) publish(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest) error {

	if req.VolumeId == "" {
		return ErrVolumeIDRequired
	}

	if req.TargetPath == "" {
		return ErrTargetPathRequired
	}

	if req.VolumeCapability == nil {
		return ErrVolumeCapabilityRequired
	}

	if req.VolumeCapability.AccessMode == nil {
		return ErrAccessModeRequired
	}

	var (
		isBlock    bool
		fsType     string
		mntFlags   []string
		readOnly   = req.Readonly
		accessMode = req.VolumeCapability.AccessMode.Mode
	)

	lf := log.Fields{
		"volumeID":   req.VolumeId,
		"targetPath": req.TargetPath,
		"readOnly":   req.Readonly,
		"accessMode": accessMode,
	}

	switch accessType := req.VolumeCapability.AccessType.(type) {
	case *csi.VolumeCapability_Block:
		if accessType.Block == nil {
			return ErrBlockTypeRequired
		}
		if readOnly {
			return status.Error(
				codes.InvalidArgument,
				"read only not supported by access type")
		}
		isBlock = true
		lf["accessType"] = "block"
	case *csi.VolumeCapability_Mount:
		if accessType.Mount == nil {
			return ErrMountTypeRequired
		}
		fsType = accessType.Mount.FsType
		mntFlags = accessType.Mount.MountFlags
		lf["accessType"] = "mount"
		lf["fsType"] = fsType
		lf["mntFlags"] = mntFlags
	default:
		return ErrAccessTypeRequired
	}

	devicePath, err := i.p.GetDevicePath(
		ctx, req.VolumeId, req.PublishVolumeInfo, req.UserCredentials)
	if err != nil {
		return err
	}
	if devicePath == "" {
		return status.Errorf(
			codes.FailedPrecondition,
			"failed to discover device path for volume: %s", req.VolumeId)
	}
	if i.validateDevice {
		validatedDevPath, err := validateDevice(devicePath)
		if err != nil {
			return status.Error(codes.FailedPrecondition, err.Error())
		}
		devicePath = validatedDevPath
	}
	lf["devicePath"] = devicePath

	privMntTgtName, err := i.getPrivateMountTargetName(
		ctx, req.VolumeId, req.UserCredentials)
	if err != nil {
		return err
	}
	privMntTgtPath := path.Join(i.privateMountDir, privMntTgtName)
	lf["privateMountTargetPath"] = privMntTgtPath

	mountTable, err := gofsutil.GetMounts(ctx)
	if err != nil {
		return err
	}

	var (
		isPub     bool
		isPrivPub bool
	)

	for _, m := range mountTable {
		// Check to see if the device is already mounted to the private
		// target path.
		if m.Source == devicePath && m.Path == privMntTgtPath {
			stat, _ := os.Stat(privMntTgtPath)
			if stat == nil {
				continue
			}
			if (isBlock && stat.IsDir()) || !stat.IsDir() {
				return fmt.Errorf(
					"invalid existing private mount target: %s",
					privMntTgtPath)
			}
			isPrivPub = true
			lf["deviceToPrivate"] = isPrivPub
		}

		// Check to see if the private target path is already mounted to the
		// target path.
		if m.Source == privMntTgtPath && m.Path == req.TargetPath {
			stat, _ := os.Stat(req.TargetPath)
			if stat == nil {
				continue
			}
			if (isBlock && stat.IsDir()) || !stat.IsDir() {
				return fmt.Errorf(
					"invalid existing target path: %s",
					req.TargetPath)
			}
			isPub = true
			lf["privateToPublic"] = isPub
		}
	}

	if !isPrivPub {
		if isBlock {
			if _, err := os.Stat(privMntTgtPath); err != nil {
				if !os.IsNotExist(err) {
					return err
				}
				log.WithFields(lf).Debug("creating private mount target file")
				f, err := os.Create(privMntTgtPath)
				if err != nil {
					return err
				}
				f.Close()
			}
			log.WithFields(lf).Debug(
				"binding mounting device to private mount target")

			if err := gofsutil.BindMount(
				ctx, devicePath, privMntTgtPath); err != nil {
				return err
			}
		} else {
			if _, err := os.Stat(privMntTgtPath); err != nil {
				if !os.IsNotExist(err) {
					return err
				}
				log.WithFields(lf).Debug("creating private mount target dir")
				if err := os.MkdirAll(privMntTgtPath, 0755); err != nil {
					return err
				}
			}
			log.WithFields(lf).Debug(
				"formatting device & mounting to private mount target")
			if f := i.p.FormatAndMount(); f != nil {
				err := f(
					ctx, devicePath, req.TargetPath, fsType,
					mntFlags, req.UserCredentials)
				if err != nil {
					return err
				}
			} else {
				err := gofsutil.FormatAndMount(
					ctx, devicePath, req.TargetPath, fsType, mntFlags...)
				if err != nil {
					return err
				}
			}
		}
	}

	if !isPub {
		log.WithFields(lf).Debug(
			"binding mounting private mount target to target path")
		err := gofsutil.BindMount(ctx, privMntTgtPath, req.TargetPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *nodeVolumePublicist) unpublish(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest) error {

	if req.VolumeId == "" {
		return ErrVolumeIDRequired
	}

	if req.TargetPath == "" {
		return ErrTargetPathRequired
	}

	lf := log.Fields{
		"volumeID":   req.VolumeId,
		"targetPath": req.TargetPath,
	}

	privMntTgtName, err := i.getPrivateMountTargetName(
		ctx, req.VolumeId, req.UserCredentials)
	if err != nil {
		return err
	}
	privMntTgtPath := path.Join(i.privateMountDir, privMntTgtName)
	lf["privateMountTargetPath"] = privMntTgtPath

	privMntTgtStat, err := os.Stat(privMntTgtPath)
	if err != nil {
		return err
	}
	isBlock := !privMntTgtStat.IsDir()
	lf["isBlock"] = isBlock

	mountTable, err := gofsutil.GetMounts(ctx)
	if err != nil {
		return err
	}

	var (
		mountPoints     int
		isTargetMounted bool
	)

	for _, m := range mountTable {
		if m.Source == privMntTgtPath {
			mountPoints++
			if m.Path == req.TargetPath {
				isTargetMounted = true
				lf["isTargetMounted"] = true
			}
		}
	}

	if isTargetMounted {
		log.WithFields(lf).Debug("unmounting target path")
		if err := gofsutil.Unmount(ctx, req.TargetPath); err != nil {
			return err
		}
	}

	// If the private mount path was only bind mounted once it means
	// that the private mount should be unmounted.
	if mountPoints == 1 {
		log.WithFields(lf).Debug("unmounting private mount target path")
		if err := gofsutil.Unmount(ctx, privMntTgtPath); err != nil {
			return err
		}
	}

	return nil
}

func (i *nodeVolumePublicist) getPrivateMountTargetName(
	ctx context.Context,
	volID string,
	userCreds map[string]string) (string, error) {

	v, err := i.p.GetPrivateMountTargetName(ctx, volID, userCreds)
	if err != nil {
		return "", err
	}
	if v != "" {
		return v, nil
	}

	h := md5.New()
	h.Write([]byte(volID))
	return hex.EncodeToString(h.Sum(nil)), nil
}

func validateDevice(devicePath string) (string, error) {

	if _, err := os.Lstat(devicePath); err != nil {
		return "", err
	}

	// Eval any symlinks to ensure the specified path points to a real device.
	realPath, err := filepath.EvalSymlinks(devicePath)
	if err != nil {
		return "", err
	}
	devicePath = realPath

	if stat, _ := os.Stat(devicePath); stat != nil ||
		stat.Mode()&os.ModeDevice == 0 {
		return "", fmt.Errorf("invalid block device: %s", devicePath)
	}

	return devicePath, nil
}
