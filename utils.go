package gocsi

import (
	"fmt"
	"net"
	"os"
	"regexp"
)

// Version is a type that responds with Major, Minor, and Patch
// information typical of a semantic version.
type Version interface {
	GetMajor() uint32
	GetMinor() uint32
	GetPatch() uint32
}

// SprintfVersion formats a Version as a string.
func SprintfVersion(v Version) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%d.%d.%d", v.GetMajor(), v.GetMinor(), v.GetPatch())
}

// CompareVersions compares two versions and returns:
//
//   -1 if a > b
//    0 if a = b
//    1 if a < b
func CompareVersions(a, b Version) int8 {
	if a == nil && b == nil {
		return 0
	}
	if a != nil && b == nil {
		return -1
	}
	if a == nil && b != nil {
		return 1
	}
	if a.GetMajor() > b.GetMajor() {
		return -1
	}
	if a.GetMajor() < b.GetMajor() {
		return 1
	}
	if a.GetMinor() > b.GetMinor() {
		return -1
	}
	if a.GetMinor() < b.GetMinor() {
		return 1
	}
	if a.GetPatch() > b.GetPatch() {
		return -1
	}
	if a.GetPatch() < b.GetPatch() {
		return 1
	}
	return 0
}

// GetCSIEndpoint returns the network address specified by the
// environment variable CSI_ENDPOINT.
func GetCSIEndpoint() (network, addr string, err error) {
	protoAddr := os.Getenv("CSI_ENDPOINT")
	if protoAddr == "" {
		return "", "", ErrMissingCSIEndpoint
	}
	return ParseProtoAddr(protoAddr)
}

// GetCSIEndpointListener returns the net.Listener for the endpoint
// specified by the environment variable CSI_ENDPOINT.
func GetCSIEndpointListener() (net.Listener, error) {
	proto, addr, err := GetCSIEndpoint()
	if err != nil {
		return nil, err
	}
	return net.Listen(proto, addr)
}

var addrRX = regexp.MustCompile(
	`(?i)^((?:(?:tcp|udp|ip)[46]?)|(?:unix(?:gram|packet)?))://(.+)$`)

// ParseProtoAddr parses a Golang network address.
func ParseProtoAddr(protoAddr string) (proto string, addr string, err error) {
	m := addrRX.FindStringSubmatch(protoAddr)
	if m == nil {
		return "", "", ErrInvalidCSIEndpoint
	}
	return m[1], m[2], nil
}
