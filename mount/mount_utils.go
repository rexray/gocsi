package mount

import (
	"bufio"
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// EntryProcessorFunc defines the signature of the function that can
// be passed to GetMounts to customize how entries in the mount table
// are handled.
//
// When validateOnly is true it's not necessary to return mount
// information, only validate that the entry is valid.
type EntryProcessorFunc func(
	ctx context.Context,
	root, mountPoint, fsType, mountSource string, mountOpts []string,
	mountSourceToRoot map[string]string,
	validateOnly bool) (Info, bool)

// GetDefaultEntryProcessor returns the default entry processor function.
func GetDefaultEntryProcessor() EntryProcessorFunc {
	return defaultEntryProcessor
}

func defaultEntryProcessor(
	ctx context.Context,
	root, mountPoint, fsType, mountSource string, mountOpts []string,
	mountSourceToMountPoint map[string]string,
	validateOnly bool) (info Info, valid bool) {

	info.Device = mountSource
	info.Path = mountPoint
	info.Type = fsType
	info.Opts = mountOpts

	// Validate the mount table entry.
	validFSType, _ := regexp.MatchString(
		`(?i)^devtmpfs|(?:fuse\..*)|(?:nfs\d?)$`, fsType)
	sourceHasSlashPrefix := strings.HasPrefix(mountSource, "/")
	if valid = validFSType || sourceHasSlashPrefix; !valid || validateOnly {
		return
	}

	// If this is the first time a source is encountered in the
	// output then cache its mountPoint field as the filesystem path
	// to which the source is mounted as a non-bind mount.
	//
	// Subsequent encounters with the source will resolve it
	// to the cached root value in order to set the mount info's
	// Source field to the the cached mountPont field value + the
	// value of the current line's root field.
	if cachedMountPoint, ok := mountSourceToMountPoint[mountSource]; ok {
		info.Source = path.Join(cachedMountPoint, root)
	} else {
		mountSourceToMountPoint[mountSource] = mountPoint
	}

	return
}

// ProcMountsFields is fields per line in procMountsPath as per
// https://www.kernel.org/doc/Documentation/filesystems/proc.txt
const ProcMountsFields = 9

/*
ReadProcMountsFrom parses the contents of a mount table file, typically
"/proc/self/mountinfo".

From https://www.kernel.org/doc/Documentation/filesystems/proc.txt:

3.5	/proc/<pid>/mountinfo - Information about mounts
--------------------------------------------------------

This file contains lines of the form:

36 35 98:0 /mnt1 /mnt2 rw,noatime master:1 - ext3 /dev/root rw,errors=continue
(1)(2)(3)   (4)   (5)      (6)      (7)   (8) (9)   (10)         (11)

(1) mount ID:  unique identifier of the mount (may be reused after umount)
(2) parent ID:  ID of parent (or of self for the top of the mount tree)
(3) major:minor:  value of st_dev for files on filesystem
(4) root:  root of the mount within the filesystem
(5) mount point:  mount point relative to the process's root
(6) mount options:  per mount options
(7) optional fields:  zero or more fields of the form "tag[:value]"
(8) separator:  marks the end of the optional fields
(9) filesystem type:  name of filesystem of the form "type[.subtype]"
(10) mount source:  filesystem specific information or "none"
(11) super options:  per super block options

Parsers should ignore all unrecognised optional fields.  Currently the
possible optional fields are:

shared:X  mount is shared in peer group X
master:X  mount is slave to peer group X
propagate_from:X  mount is slave and receives propagation from peer group X (*)
unbindable  mount is unbindable
*/
func ReadProcMountsFrom(
	ctx context.Context,
	file io.Reader,
	quick bool,
	expectedFields int,
	processor EntryProcessorFunc) ([]Info, uint32, error) {

	if processor == nil {
		processor = defaultEntryProcessor
	}

	var (
		mountInfos              []Info
		mountSourceToMountPoint map[string]string
	)

	if !quick {
		mountSourceToMountPoint = map[string]string{}
	}

	hash := fnv.New32a()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {

		line := scanner.Text()
		fields := strings.Fields(line)

		// Remove the optional fields that should be ignored.
		for {
			val := fields[6]
			fields = append(fields[:6], fields[7:]...)
			if val == "-" {
				break
			}
		}

		if len(fields) != expectedFields {
			return nil, 0, fmt.Errorf(
				"readProcMountsFrom: invalid field count: exp=%d, act=%d: %s",
				expectedFields, len(fields), line)
		}

		info, valid := processor(
			ctx,
			fields[3],                     // root
			fields[4],                     // mountPoint
			fields[6],                     // fsType
			fields[7],                     // mountSource
			strings.Split(fields[5], ","), // mountOpts
			mountSourceToMountPoint,
			quick)

		if !valid {
			continue
		}

		fmt.Fprint(hash, line)

		if quick {
			continue
		}

		mountInfos = append(mountInfos, info)
	}
	return mountInfos, hash.Sum32(), nil
}

// EvalSymlinks evaluates the provided path and updates it to remove
// any symlinks in its structure, replacing them with the actual path
// components.
func EvalSymlinks(symPath *string) error {
	realPath, err := filepath.EvalSymlinks(*symPath)
	if err != nil {
		return err
	}
	*symPath = realPath
	return nil
}
