# CSI Language Bindings for Go
This package contains the CSI language bindings for Go, generated using
the protobuf compiler `protoc`. GoCSI references the
[CSI spec](https://github.com/container-storage-interface/spec)
project in order to obtain the CSI specification.

## Updating the Specification Version
In order to update the CSI specification version used by GoCSI, please
follow the steps below starting from the root of the GoCSI project:

1. Update `glide.yaml` with the desired CSI specification version.
2. Run `make glide-up` to update the vendored dependencies.
3. Run `make test` to update the generated protobuf source and execute
the test suite using the Mock plug-in.

If all of the above steps complete without error then commit and push
the changes and use GitHub to create a pull request.
