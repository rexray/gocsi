package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/thecodeteam/gocsi"
	"github.com/thecodeteam/gocsi/csi"
)

const userCredsKey = "X_CSI_USER_CREDENTIALS"

// mapOfStringArg is used for parsing a csv, key=value arg into
// a map[string]string
type mapOfStringArg struct {
	sync.Once
	val  string
	data map[string]string
}

func (s *mapOfStringArg) String() string {
	return s.val
}

func (s *mapOfStringArg) Type() string {
	return "map[string]string"
}

func (s *mapOfStringArg) Set(val string) error {
	s.Do(func() { s.data = map[string]string{} })
	data := strings.Split(val, ",")
	for _, v := range data {
		vp := strings.SplitN(v, "=", 2)
		switch len(vp) {
		case 1:
			s.data[vp[0]] = ""
		case 2:
			s.data[vp[0]] = vp[1]
		}
	}
	return nil
}

// csiVersionArg is used for parsing a CSI version argument
type csiVersionArg struct {
	csi.Version
	val string
}

func (s *csiVersionArg) String() string {
	return s.val
}

func (s *csiVersionArg) Type() string {
	return "csi.Version"
}

func (s *csiVersionArg) Set(val string) error {
	var err error
	if s.Version, err = gocsi.ParseVersion(val); err != nil {
		return err
	}
	return nil
}

// volumeCapabilitySliceArg is used for parsing one or more volume
// capabilities from the command line
type volumeCapabilitySliceArg struct {
	val  string
	data []*csi.VolumeCapability
}

func (s *volumeCapabilitySliceArg) String() string {
	return s.val
}

func (s *volumeCapabilitySliceArg) Type() string {
	return "csi.VolumeCapability"
}

func (s *volumeCapabilitySliceArg) Set(val string) error {
	// The data can be split into a max of 4 components:
	// 1. mode
	// 2. cap
	// 3. fsType (if cap is mount)
	// 4. mntFlags (if cap is mount)
	data := strings.SplitN(val, ",", 4)
	if len(data) < 2 {
		return fmt.Errorf("invalid volume capability: %s", val)
	}

	var cap csi.VolumeCapability

	szMode := data[0]
	if i, ok := csi.VolumeCapability_AccessMode_Mode_value[szMode]; ok {
		cap.AccessMode = &csi.VolumeCapability_AccessMode{
			Mode: csi.VolumeCapability_AccessMode_Mode(i),
		}
	} else {
		i, err := strconv.ParseInt(szMode, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid access mode: %v: %v", szMode, err)
		}
		if _, ok := csi.VolumeCapability_AccessMode_Mode_name[int32(i)]; ok {
			cap.AccessMode = &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_Mode(i),
			}
		}
	}

	szType := data[1]

	// Handle a block volume capability
	if szType == "1" || strings.EqualFold(szType, "block") {
		cap.AccessType = &csi.VolumeCapability_Block{
			Block: &csi.VolumeCapability_BlockVolume{},
		}
		s.data = append(s.data, &cap)
		return nil
	}

	// Handle a mount volume capability
	if szType == "2" || strings.EqualFold(szType, "mount") {
		if len(data) < 3 {
			return fmt.Errorf("invalid volume capability: %s", val)
		}
		mountCap := &csi.VolumeCapability_MountVolume{
			FsType: data[2],
		}
		cap.AccessType = &csi.VolumeCapability_Mount{
			Mount: mountCap,
		}

		// If there is data remaining then treat it as mount flags.
		if len(data) > 3 {
			mountCap.MountFlags = strings.Split(data[3], ",")
		}

		s.data = append(s.data, &cap)
		return nil
	}

	return fmt.Errorf("invalid volume capability: %s", val)
}
