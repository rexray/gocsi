package mount

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

var (
	bindRemountOpts = []string{}
	mountRX         = regexp.MustCompile(`^(.+) on (.+) \((.+)\)$`)
)

// getDiskFormat uses 'lsblk' to see if the given disk is unformated
func getDiskFormat(
	ctx context.Context,
	disk string,
	processor EntryProcessorFunc) (string, error) {

	mps, err := getMounts(ctx, processor)
	if err != nil {
		return "", err
	}
	for _, i := range mps {
		if i.Device == disk {
			return i.Type, nil
		}
	}
	return "", fmt.Errorf("getDiskFormat: failed: %s", disk)
}

// formatAndMount uses unix utils to format and mount the given disk
func formatAndMount(source, target, fsType string, options []string) error {
	return ErrNotImplemented
}

// getMounts returns a slice of all the mounted filesystems
func getMounts(
	ctx context.Context,
	processor EntryProcessorFunc) ([]Info, error) {

	out, err := exec.Command("mount").CombinedOutput()
	if err != nil {
		return nil, err
	}

	var mountInfos []Info
	scan := bufio.NewScanner(bytes.NewReader(out))

	for scan.Scan() {
		m := mountRX.FindStringSubmatch(scan.Text())
		if len(m) != 4 {
			continue
		}
		device := m[1]
		if !strings.HasPrefix(device, "/") {
			continue
		}
		var (
			path    = m[2]
			source  = device
			options = strings.Split(m[3], ",")
		)
		if len(options) == 0 {
			return nil, fmt.Errorf(
				"getMounts: invalid mount options: %s", device)
		}
		for i, v := range options {
			options[i] = strings.TrimSpace(v)
		}
		fsType := options[0]
		if len(options) > 1 {
			options = options[1:]
		} else {
			options = nil
		}
		mountInfos = append(mountInfos, Info{
			Device: device,
			Path:   path,
			Source: source,
			Type:   fsType,
			Opts:   options,
		})
	}
	return mountInfos, nil
}

// bindMount performs a bind mount
func bindMount(source, target string, options []string) error {
	return doMount("bindfs", source, target, "", options)
}
