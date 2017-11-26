package main

import (
	"context"

	"github.com/thecodeteam/gocsi/csp"
	"github.com/thecodeteam/gocsi/mock/provider"
	"github.com/thecodeteam/gocsi/mock/service"
)

// main is ignored when this package is built as a go plug-in
func main() {
	csp.Run(
		context.Background(),
		service.Name,
		"A Mock Container Storage Interface (CSI) Storage Plug-in (SP)",
		"",
		provider.New())
}
