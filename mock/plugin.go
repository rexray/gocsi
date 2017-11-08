// +build linux,plugin

package main

import "C"

import (
	"fmt"

	"github.com/thecodeteam/gocsi/mock/provider"
	"github.com/thecodeteam/gocsi/mock/service"
)

////////////////////////////////////////////////////////////////////////////////
//                              Go Plug-in                                    //
////////////////////////////////////////////////////////////////////////////////

func init() {
	SupportedVersions := make([]string, len(service.SupportedVersions))
	for i, v := range service.SupportedVersions {
		SupportedVersions[i] = fmt.Sprintf(
			"%d.%d.%d", v.Major, v.Minor, v.Patch)
	}
}

// SupportedVersions is a list of supported CSI versions as string values.
var SupportedVersions []string

// ServiceProviders is an exported symbol that provides a host program
// with a map of the service provider names and constructors.
var ServiceProviders = map[string]func() interface{}{
	service.Name: func() { return provider.New() },
}
