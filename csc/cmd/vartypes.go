package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

// mapOfStringArg is used for parsing a csv, key=value arg into
// a map[string]string
type mapOfStringArg struct {
	sync.Once
	data map[string]string
}

func (s *mapOfStringArg) String() string {
	return ""
}

func (s *mapOfStringArg) Type() string {
	return "key=val[,key=val,...]"
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

// volumeCapabilitySliceArg is used for parsing one or more volume
// capabilities from the command line
type volumeCapabilitySliceArg struct {
	data []*csi.VolumeCapability
}

func (s *volumeCapabilitySliceArg) String() string {
	return ""
}

func (s *volumeCapabilitySliceArg) Type() string {
	return "mode,type[,fstype,mntflags]"
}

func (s *volumeCapabilitySliceArg) Set(val string) error {
	// The data can be split into a max of 4 components:
	// 1. mode
	// 2. volCap
	// 3. fsType (if volCap is mount)
	// 4. mntFlags (if volCap is mount)
	data := strings.SplitN(val, ",", 4)
	if len(data) < 2 {
		return fmt.Errorf("invalid volume capability: %s", val)
	}

	var volCap csi.VolumeCapability

	szMode := data[0]
	if i, ok := csi.VolumeCapability_AccessMode_Mode_value[szMode]; ok {
		volCap.AccessMode = &csi.VolumeCapability_AccessMode{
			Mode: csi.VolumeCapability_AccessMode_Mode(i),
		}
	} else {
		i, err := strconv.ParseInt(szMode, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid access mode: %v: %v", szMode, err)
		}
		if _, ok := csi.VolumeCapability_AccessMode_Mode_name[int32(i)]; ok {
			volCap.AccessMode = &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_Mode(i),
			}
		}
	}

	szType := data[1]

	// Handle a block volume capability
	if szType == "1" || strings.EqualFold(szType, "block") {
		volCap.AccessType = &csi.VolumeCapability_Block{
			Block: &csi.VolumeCapability_BlockVolume{},
		}
		s.data = append(s.data, &volCap)
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
		volCap.AccessType = &csi.VolumeCapability_Mount{
			Mount: mountCap,
		}

		// If there is data remaining then treat it as mount flags.
		if len(data) > 3 {
			mountCap.MountFlags = strings.Split(data[3], ",")
		}

		s.data = append(s.data, &volCap)
		return nil
	}

	return fmt.Errorf("invalid volume capability: %s", val)
}

// docTypeArg is used for parsing the doc type to generate
type docTypeArg struct {
	val string
}

func (s *docTypeArg) String() string {
	return "md"
}

func (s *docTypeArg) Type() string {
	return "md|man|rst"
}

func (s *docTypeArg) Set(val string) error {
	val = strings.ToLower(val)
	switch val {
	case "md":
		s.val = val
		return nil
	case "man":
		s.val = val
		return nil
	case "rst":
		s.val = val
		return nil
	}
	return fmt.Errorf("invalid doc type: %s", val)
}

type logLevelArg struct {
	pflag.Value
	val log.Level
	set bool
}

func (a *logLevelArg) Val() (log.Level, bool) {
	if a.set {
		return a.val, false
	}
	return log.WarnLevel, true
}

func (a *logLevelArg) String() string {
	if a.set {
		return a.val.String()
	}
	return "WARN"
}

func (a *logLevelArg) Type() string {
	return "PANIC|FATAL|ERROR|WARN|INFO|DEBUG"
}

func (a *logLevelArg) Set(val string) error {
	lvl, err := log.ParseLevel(val)
	if err != nil {
		return err
	}
	a.val = lvl
	a.set = true
	return nil
}
