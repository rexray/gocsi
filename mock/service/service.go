package service

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/thecodeteam/gocsi"
	"github.com/container-storage-interface/spec/lib/go/csi"
)

const (
	// Name is the name of the CSI plug-in.
	Name = "csi-mock"

	// VendorVersion is the version returned by GetPluginInfo.
	VendorVersion = "0.1.0"

	// SupportedVersions is a list of supported CSI versions.
	SupportedVersions = "0.1.0 0.2.0 1.0.0 1.1.0"
)

// Service is the CSI Mock service provider.
type Service interface {
	csi.ControllerServer
	csi.IdentityServer
	csi.NodeServer
	gocsi.IdempotencyProvider
}

type service struct {
	sync.Mutex
	nodeID  string
	vols    []csi.VolumeInfo
	volsRWL sync.RWMutex
	volsNID uint64
}

// New returns a new Service.
func New() Service {
	s := &service{nodeID: Name}
	s.vols = []csi.VolumeInfo{
		s.newVolume("Mock Volume 1", gib100),
		s.newVolume("Mock Volume 2", gib100),
		s.newVolume("Mock Volume 3", gib100),
	}
	return s
}

const (
	kib    uint64 = 1024
	mib    uint64 = kib * 1024
	gib    uint64 = mib * 1024
	gib100 uint64 = gib * 100
	tib    uint64 = gib * 1024
	tib100 uint64 = tib * 100
)

var version = &csi.Version{Major: 0, Minor: 1, Patch: 0}

func (s *service) newVolume(name string, capcity uint64) csi.VolumeInfo {
	return csi.VolumeInfo{
		Id:            fmt.Sprintf("%d", atomic.AddUint64(&s.volsNID, 1)),
		Attributes:    map[string]string{"name": name},
		CapacityBytes: capcity,
	}
}

func (s *service) findVol(k, v string) (volIdx int, volInfo csi.VolumeInfo) {
	s.volsRWL.RLock()
	defer s.volsRWL.RUnlock()
	return s.findVolNoLock(k, v)
}

func (s *service) findVolNoLock(k, v string) (volIdx int, volInfo csi.VolumeInfo) {
	volIdx = -1

	for i, vi := range s.vols {
		switch k {
		case "id":
			if strings.EqualFold(v, vi.Id) {
				return i, vi
			}
		case "name":
			if n, ok := vi.Attributes["name"]; ok && strings.EqualFold(v, n) {
				return i, vi
			}
		}
	}

	return
}
