# GoCSI
A Container Storage Interface (CSI) library, client, and other helpful
utilities created with Go.

## CSI Specification Version
GoCSI references the
[CSI spec](https://github.com/container-storage-interface/spec)
project in order to obtain the CSI specification. To update the version
of the specification used by GoCSI to generate language bindings, please
update the `glide.yaml`, execute `make glide-up`, and finally, please
run `make` to rebuild the Go language bindings from the updated
specification.

## Build Reference
All of GoCSI's dependency's are vendored, so GoCSI is go gettable with
`go get github.com/codedellemc/gocsi`. However, if GoCSI is referenced
by another project it is recommended that the project strip GoCSI's
`vendor` directory and supply the contained dependencies directly.
