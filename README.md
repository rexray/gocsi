# GoCSI
The Container Storage Interface
([CSI](https://github.com/container-storage-interface/spec))
is an industry standard specification for creating storage plug-ins
for container orchestrators. GoCSI aids in the development and testing
of CSI plug-ins:

| Component | Description |
|-----------|-------------|
| [gocsi](#library) | CSI Go library |
| [csc](./csc/) | CSI command line interface (CLI) client |
| [csp](./csp) | CSI storage plug-in (CSP) bootstrapper |
| [mock](./mock) | CSI mock storage plug-in (SP) |

## Quick Start
The following example illustrates using Docker in combination with the
GoCSI SP bootstrapper [csp](./csp) to create a new CSI SP from scratch,
serve it on a UNIX socket, and then use the GoCSI command line client
[`csc`](./csc/) to invoke the `GetSupportedVersions` and `GetPluginInfo`
RPCs:

```shell
$ docker run -it golang:latest sh -c \
  "go get github.com/thecodeteam/gocsi && \
  make -C src/github.com/thecodeteam/gocsi csi-sp"
```

## Library
The root of the GoCSI project is a general purpose library for CSI. This
package provides the following features:

* [gRPC interceptors](#interceptors)
* A [channel-based variant](#pagevolumes) of `ListVolumes`

### Interceptors
GoCSI includes the following gRPC client-side and server-side interceptors:

| Type | Client | Server | Description |
|------|--------|--------|-------------|
| Request & response logging | ✓ | ✓ | Logs request & response data (except `UserCredentials`) |
| Request ID injector | ✓ | ✓ | Injects outgoing (or incoming) requests with a unique ID |
| Spec validator | ✓ | ✓ | Validates requests & responses against the CSI spec |
| Idempotency | | ✓ | Assists in making an SP idempotent |

Please refer to the CSI client [`csc`](./csc/cmd/interceptors.go) for
examples of how to implement the client-side interceptors. The
[`csp` package](./csp/csp_interceptors.go) illustrates the use of
GoCSI's server-side interceptors.

### `PageVolumes`
The `PageVolumes` function invokes the `ListVolumes` RPC until all
available volumes are retrieved, returning them over a Go channel.

```go
func PageVolumes(
	ctx context.Context,
	client csi.ControllerClient,
	req csi.ListVolumesRequest,
	opts ...grpc.CallOption) (<-chan csi.VolumeInfo, <-chan error)
```

The `csc` command `controller listvolumes --paging`
[uses `PageVolumes`](./csc/cmd/controller_list_volumes.go#L43)
to stream volumes from an SP in order to minimize the amount of memory
required for a client to process all available volumes.
