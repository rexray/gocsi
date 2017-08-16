//go:generate make

// Package gocsi provides a Container Storage Interface (CSI) library,
// client, and other helpful utilities.
package gocsi

import (
	// Always load the CSI package.
	_ "github.com/codedellemc/gocsi/csi"
)
