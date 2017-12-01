package gofsutil_test

import (
	"context"
	"path"
	"regexp"
	"strings"
	"testing"

	"github.com/thecodeteam/gofsutil"
)

func newTestEntryScanFunc(t *testing.T) gofsutil.EntryScanFunc {
	//return gofsutil.DefaultEntryScanFunc()
	return (&testEntryScanFunc{t}).scan
}

type testEntryScanFunc struct {
	t *testing.T
}

func (p *testEntryScanFunc) scan(
	ctx context.Context,
	entry gofsutil.Entry,
	cache map[string]gofsutil.Entry) (
	info gofsutil.Info, valid bool, failed error) {

	// p.t.Logf("root=%s\tmountPoint=%s\t"+
	// 	"fsType=%s\tmountSource=%s\tmountOpts=%v",
	// 	root, mountPoint, fsType, mountSource, mountOpts)

	baseName := entry.Root
	if isNFS, _ := regexp.MatchString(`(?i)^nfs\d?$`, entry.FSType); isNFS {
		baseName = path.Base(entry.MountSource)
		entry.MountSource = path.Dir(entry.MountSource)
	}

	// Validate the mount table entry.
	validFSType, _ := regexp.MatchString(
		`(?i)^devtmpfs|(?:fuse\..*)|(?:nfs\d?)$`, entry.FSType)
	sourceHasSlashPrefix := strings.HasPrefix(entry.MountSource, "/")
	if valid = validFSType || sourceHasSlashPrefix; !valid {
		return
	}

	// Copy the Entry object's fields to the Info object.
	info.Device = entry.MountSource
	info.Opts = make([]string, len(entry.MountOpts))
	copy(info.Opts, entry.MountOpts)
	info.Path = entry.MountPoint
	info.Type = entry.FSType
	info.Source = entry.MountSource

	// If this is the first time a source is encountered in the
	// output then cache its mountPoint field as the filesystem path
	// to which the source is mounted as a non-bind mount.
	//
	// Subsequent encounters with the source will resolve it
	// to the cached root value in order to set the mount info's
	// Source field to the the cached mountPont field value + the
	// value of the current line's root field.
	if cachedEntry, ok := cache[entry.MountSource]; ok {
		info.Source = path.Join(cachedEntry.MountPoint, baseName)
	} else {
		cache[entry.MountSource] = entry
	}

	return
}

func TestReadProcMountsFrom(t *testing.T) {

	mountInfos, _, err := gofsutil.ReadProcMountsFrom(
		context.TODO(),
		strings.NewReader(procMountInfoData),
		false,
		gofsutil.ProcMountsFields,
		newTestEntryScanFunc(t))
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("len(mounts)=%d", len(mountInfos))
	success1 := "/home/akutz/2"
	success2 := "/home/akutz/travis-is-right"
	success3 := "/home/akutz/red"
	success4 := "/var/lib/rexray/volumes/s3fsvol01"
	for _, mi := range mountInfos {
		t.Logf("%+v", mi)
		if mi.Path == "/home/akutz/2" && mi.Source == "/home/akutz/1" {
			success1 = ""
		}
		if mi.Path == "/home/akutz/travis-is-right" && mi.Source == "/dev/sda1" {
			success2 = ""
		}
		if mi.Path == "/home/akutz/red" && mi.Device == "localhost:/home" {
			success3 = ""
		}
		if mi.Path == "/var/lib/rexray/volumes/s3fsvol01" && mi.Device == "s3fs" {
			success4 = ""
		}
	}

	chk := func(s string) {
		if s != "" {
			t.Errorf("error: %s", s)
			t.Fail()
		}
	}

	chk(success1)
	chk(success2)
	chk(success3)
	chk(success4)
}

const procMountInfoData = `17 60 0:16 / /sys rw,nosuid,nodev,noexec,relatime shared:6 - sysfs sysfs rw,seclabel
18 60 0:3 / /proc rw,nosuid,nodev,noexec,relatime shared:5 - proc proc rw
19 60 0:5 / /dev rw,nosuid shared:2 - devtmpfs devtmpfs rw,seclabel,size=1930460k,nr_inodes=482615,mode=755
20 17 0:15 / /sys/kernel/security rw,nosuid,nodev,noexec,relatime shared:7 - securityfs securityfs rw
21 19 0:17 / /dev/shm rw,nosuid,nodev shared:3 - tmpfs tmpfs rw,seclabel
22 19 0:11 / /dev/pts rw,nosuid,noexec,relatime shared:4 - devpts devpts rw,seclabel,gid=5,mode=620,ptmxmode=000
23 60 0:18 / /run rw,nosuid,nodev shared:23 - tmpfs tmpfs rw,seclabel,mode=755
24 17 0:19 / /sys/fs/cgroup ro,nosuid,nodev,noexec shared:8 - tmpfs tmpfs ro,seclabel,mode=755
25 24 0:20 / /sys/fs/cgroup/systemd rw,nosuid,nodev,noexec,relatime shared:9 - cgroup cgroup rw,xattr,release_agent=/usr/lib/systemd/systemd-cgroups-agent,name=systemd
26 17 0:21 / /sys/fs/pstore rw,nosuid,nodev,noexec,relatime shared:20 - pstore pstore rw
27 24 0:22 / /sys/fs/cgroup/cpu,cpuacct rw,nosuid,nodev,noexec,relatime shared:10 - cgroup cgroup rw,cpuacct,cpu
28 24 0:23 / /sys/fs/cgroup/hugetlb rw,nosuid,nodev,noexec,relatime shared:11 - cgroup cgroup rw,hugetlb
29 24 0:24 / /sys/fs/cgroup/perf_event rw,nosuid,nodev,noexec,relatime shared:12 - cgroup cgroup rw,perf_event
30 24 0:25 / /sys/fs/cgroup/net_cls,net_prio rw,nosuid,nodev,noexec,relatime shared:13 - cgroup cgroup rw,net_prio,net_cls
31 24 0:26 / /sys/fs/cgroup/blkio rw,nosuid,nodev,noexec,relatime shared:14 - cgroup cgroup rw,blkio
32 24 0:27 / /sys/fs/cgroup/devices rw,nosuid,nodev,noexec,relatime shared:15 - cgroup cgroup rw,devices
33 24 0:28 / /sys/fs/cgroup/pids rw,nosuid,nodev,noexec,relatime shared:16 - cgroup cgroup rw,pids
34 24 0:29 / /sys/fs/cgroup/freezer rw,nosuid,nodev,noexec,relatime shared:17 - cgroup cgroup rw,freezer
35 24 0:30 / /sys/fs/cgroup/memory rw,nosuid,nodev,noexec,relatime shared:18 - cgroup cgroup rw,memory
36 24 0:31 / /sys/fs/cgroup/cpuset rw,nosuid,nodev,noexec,relatime shared:19 - cgroup cgroup rw,cpuset
58 17 0:32 / /sys/kernel/config rw,relatime shared:21 - configfs configfs rw
60 1 253:0 / / rw,relatime shared:1 - xfs /dev/mapper/cl-root rw,seclabel,attr2,inode64,noquota
37 17 0:14 / /sys/fs/selinux rw,relatime shared:22 - selinuxfs selinuxfs rw
38 18 0:33 / /proc/sys/fs/binfmt_misc rw,relatime shared:24 - autofs systemd-1 rw,fd=25,pgrp=1,timeout=300,minproto=5,maxproto=5,direct
39 17 0:6 / /sys/kernel/debug rw,relatime shared:25 - debugfs debugfs rw
40 19 0:34 / /dev/hugepages rw,relatime shared:26 - hugetlbfs hugetlbfs rw,seclabel
41 19 0:13 / /dev/mqueue rw,relatime shared:27 - mqueue mqueue rw,seclabel
72 60 8:1 / /boot rw,relatime shared:28 - xfs /dev/sda1 rw,seclabel,attr2,inode64,noquota
74 60 253:2 / /home rw,relatime shared:29 - xfs /dev/mapper/cl-home rw,seclabel,attr2,inode64,noquota
150 60 253:0 /var/lib/docker/devicemapper /var/lib/docker/devicemapper rw,relatime - xfs /dev/mapper/cl-root rw,seclabel,attr2,inode64,noquota
109 23 0:35 / /run/user/1000 rw,nosuid,nodev,relatime shared:62 - tmpfs tmpfs rw,seclabel,size=388200k,mode=700,uid=1000,gid=1000
116 38 0:36 / /proc/sys/fs/binfmt_misc rw,relatime shared:66 - binfmt_misc binfmt_misc rw
113 17 0:37 / /sys/fs/fuse/connections rw,relatime shared:65 - fusectl fusectl rw
119 74 253:2 /akutz/1 /home/akutz/2 rw,relatime shared:29 - xfs /dev/mapper/cl-home rw,seclabel,attr2,inode64,noquota
119 74 0:5 /sda1 /home/akutz/travis-is-right rw,nosuid shared:2 - devtmpfs devtmpfs rw,seclabel,size=1930460k,nr_inodes=482615,mode=755
125 18 0:39 / /proc/fs/nfsd rw,relatime shared:72 - nfsd nfsd rw
128 74 0:41 / /home/akutz/red rw,relatime shared:74 - nfs4 localhost:/home/akutz rw,vers=4.1,rsize=524288,wsize=524288,namlen=255,hard,proto=tcp6,port=0,timeo=600,retrans=2,sec=sys,clientaddr=::1,local_lock=none,addr=::1
81 62 0:39 / /var/lib/rexray/volumes/s3fsvol01 rw,nosuid,nodev,relatime shared:31 - fuse.s3fs s3fs rw,user_id=0,group_id=0
121 61 0:39 / /var/lib/rexray/volumes/vol01 rw,relatime shared:69 - nfs 192.168.1.80:/ifs/vols/vol01 rw,vers=3,rsize=131072,wsize=524288,namlen=255,hard,proto=tcp,timeo=600,retrans=2,sec=sys,mountaddr=192.168.1.80,mountvers=3,mountport=300,mountproto=udp,local_lock=none,addr=192.168.1.80
124 61 0:39 / /var/lib/rexray/csi/volumes/vol01 rw,relatime shared:69 - nfs 192.168.1.80:/ifs/vols/vol01/data rw,vers=3,rsize=131072,wsize=524288,namlen=255,hard,proto=tcp,timeo=600,retrans=2,sec=sys,mountaddr=192.168.1.80,mountvers=3,mountport=300,mountproto=udp,local_lock=none,addr=192.168.1.80
`
