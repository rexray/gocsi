//go:generate make

// Package gocsi provides a Container Storage Interface (CSI) library,
// client, and other helpful utilities.
package gocsi

import (
	// Always load the CSI package.
	_ "github.com/codedellemc/gocsi/csi"
)

const (
	// Namespace is the namesapce used by the protobuf.
	Namespace = "csi"

	// CSIEndpoint is the name of the environment variable that
	// contains the CSI endpoint.
	CSIEndpoint = "CSI_ENDPOINT"
)
