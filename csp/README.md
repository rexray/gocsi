# Container Storage Plug-in (CSP)
The CSP package enables the rapid development of Container Storage Interface
(CSI) storage plug-ins (SP).

## Overview
Creating a new CSI SP using the CSP is simple and straight-forward. The
GoCSI [Mock SP](../mock) illustrates the features and configuration options
available to a CSI SP by leveraging CSP.

## Creating a Storage Plug-in
The script `csp.sh` can be used to create a new CSI SP:

```
$ ./csp.sh
usage: ./csp.sh GO_IMPORT_PATH
```

The following example demonstrates creating a new SP at the Go import
path `github.com/thecodeteam/csi-sp`:

```shell
$ ./csp.sh github.com/thecodeteam/csi-sp
creating project directories:
  /home/akutz/go/src/github.com/thecodeteam/csi-sp
  /home/akutz/go/src/github.com/thecodeteam/csi-sp/provider
  /home/akutz/go/src/github.com/thecodeteam/csi-sp/service
creating project files:
  /home/akutz/go/src/github.com/thecodeteam/csi-sp/main.go
  /home/akutz/go/src/github.com/thecodeteam/csi-sp/provider/provider.go
  /home/akutz/go/src/github.com/thecodeteam/csi-sp/service/service.go
  /home/akutz/go/src/github.com/thecodeteam/csi-sp/service/controller.go
  /home/akutz/go/src/github.com/thecodeteam/csi-sp/service/idemp.go
  /home/akutz/go/src/github.com/thecodeteam/csi-sp/service/identity.go
  /home/akutz/go/src/github.com/thecodeteam/csi-sp/service/node.go
use golang/dep? Enter yes (default) or no and press [ENTER]:
  downloading golang/dep@v0.3.2
  executing dep init
building csi-sp:
  success!
  example: CSI_ENDPOINT=csi.sock \
           /home/akutz/go/src/github.com/thecodeteam/csi-sp/csi-sp
```

The new SP adheres to the following structure:

```
|-- provider
|   |
|   |-- provider.go
|
|-- service
|   |
|   |-- controller.go
|   |-- idemp.go
|   |-- identity.go
|   |-- node.go
|   |-- service.go
|
|-- main.go
```

### Provider
The `provider` package leverages `csp` to construct an SP from the CSI
services defined in `services` package. The file `provider.go` may be
modified to:

* Add an idempotency provider
* Supply default values for the SP's environment variable configuration properties

The generated file configures the following options and their default values:

| Option | Value | Description |
|--------|-------|-------------|
| `X_CSI_IDEMP` | `true` | In concert with the file `service/idemp.go`, this option enables idempotency for the SP |
| `X_CSI_SUPPORTED_VERSIONS` | `0.0.0` | The CSI versions supported by the SP. Settings this option also relieves the SP of its responsibility to provide an implementation of the RPC `GetSupportedVersions` |

Please see the Mock SP's [`provider.go`](../mock/provider/provider.go) file
for a more complete example.

### Service
The `service` package is where the business logic occurs. The files `controller.go`,
`identity.go`, and `node.go` each correspond to their eponymous CSI services. A
developer creating a new CSI SP with `csp` will work mostly in these files. Each
of the files have a complete skeleton implementation for their respective service's
remote procedure calls (RPC).

The file `idemp.go` contains a skeleton implementation of the interface
[`gocsi.IdemptotencyProvider`](../interceptors_idempotency.go). An SP is able
to achieve idempotency simply by implementing the functions contained in this
file.

### Main
The root, or `main`, package leverages `csp` to launch the SP as a stand-alone
server process. The only requirement is that the environment variable `CSI_ENDPOINT`
must be set, otherwise a help screen is emitted that lists all of the SP's available
configuration options (environment variables).

## Configuration
All CSI SPs created using this package are able to leverage the following
environment variables:

<table>
  <thead>
    <tr>
      <th>Name</th>
      <th>Description</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><code>CSI_ENDPOINT</code></td>
      <td>
        <p>The CSI endpoint may also be specified by the environment variable
        CSI_ENDPOINT. The endpoint should adhere to Go's network address
        pattern:</p>
        <ul>
          <li><code>tcp://host:port</code></li>
          <li><code>unix:///path/to/file.sock</code></li>
        </ul>
        <p>If the network type is omitted then the value is assumed to be an
        absolute or relative filesystem path to a UNIX socket file.</p>
      </td>
    </tr>
    <tr>
      <td><code>X_CSI_DEBUG</code></td>
      <td>A <code>true</code> value is equivalent to:
        <ul>
          <li><code>X_CSI_LOG_LEVEL=debug</code></li>
          <li><code>X_CSI_REQ_LOGGING=true</code></li>
          <li><code>X_CSI_REP_LOGGING=true</code></li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><code>X_CSI_LOG_LEVEL</code></td>
      <td>
        <p>The log level. Valid values include:</p>
        <ul>
          <li><code>PANIC</code></li>
          <li><code>FATAL</code></li>
          <li><code>ERROR</code></li>
          <li><code>WARN</code></li>
          <li><code>INFO</code></li>
          <li><code>DEBUG</code></li>
        </ul>
        <p>The default value is <code>WARN</code>.</p>
      </td>
    </tr>
    <tr>
      <td><code>X_CSI_SUPPORTED_VERSIONS</code></td>
      <td>A space-delimited list of versions formatted
      <code>MAJOR.MINOR.PATCH.</code> Setting this environment variable
      bypasses the SP's <code>GetSupportedVersions</code> RPC and returns
      this list of versions instead.</td>
    </tr>
    <tr>
      <td><code>X_CSI_REQ_LOGGING</code></td>
      <td><p>A flag that enables logging of incoming requests to
      <code>STDOUT</code>.</p>
      <p>Enabling this option sets <code>X_CSI_REQ_ID_INJECTION=true</code>.</p>
      </td>
    </tr>
    <tr>
      <td><code>X_CSI_REP_LOGGING</code></td>
      <td><p>A flag that enables logging of incoming responses to
      <code>STDOUT</code>.</p>
      <p>Enabling this option sets <code>X_CSI_REQ_ID_INJECTION=true</code>.</p>
      </td>
    </tr>
    <tr>
      <td><code>X_CSI_REQ_ID_INJECTION</code></td>
      <td>A flag that enables request ID injection. The ID is parsed from
      the incoming request's metadata with a key of
      <code>csi.requestid</code>.
      If no value for that key is found then a new request ID is
      generated using an atomic sequence counter.</td>
    </tr>
    <tr>
      <td><code>X_CSI_SPEC_VALIDATION</code></td>
      <td>A flag that enables validation of incoming requests and outgoing
      responses against the CSI specification.</td>
    </tr>
    <tr>
      <td><code>X_CSI_CREATE_VOL_ALREADY_EXISTS</code></td>
      <td><p>A flag that enables treating <code>CreateVolume</code> responses
      as successful when they have an associated error code of
      <code>AlreadyExists</code>.</p>
      <p>Enabling this option sets <code>X_CSI_SPEC_VALIDATION=true</code></p>
      </td>
    </tr>
    <tr>
      <td><code>X_CSI_DELETE_VOL_NOT_FOUND</code></td>
      <td><p>A flag that enables treating <code>DeleteVolume</code> responses
      as successful when they have an associated error code of
      <code>NotFound</code>.</p>
      <p>Enabling this option sets <code>X_CSI_SPEC_VALIDATION=true</code></p>
      </td>
    </tr>
    <tr>
      <td><code>X_CSI_REQUIRE_NODE_ID</code></td>
      <td>
        <p>A flag that enables treating the following fields as required:</p>
        <ul>
          <li><code>ControllerPublishVolumeRequest.NodeId</code></li>
          <li><code>GetNodeIDResponse.NodeId</code></li>
      </ul>
      <p>Enabling this option sets <code>X_CSI_SPEC_VALIDATION=true</code></p>
      </td>
    </tr>
    <tr>
      <td><code>X_CSI_REQUIRE_PUB_VOL_INFO</code></td>
      <td>
        <p>A flag that enables treating the following fields as required:</p>
        <ul>
          <li><code>ControllerPublishVolumeResponse.PublishVolumeInfo</code></li>
          <li><code>NodePublishVolumeRequest.PublishVolumeInfo</code></li>
        </ul>
        <p>Enabling this option sets <code>X_CSI_SPEC_VALIDATION=true</code></p>
      </td>
    </tr>
    <tr>
      <td><code>X_CSI_REQUIRE_VOL_ATTRIBS</code></td>
      <td>
        <p>A flag that enables treating the following fields as required:</p>
        <ul>
          <li><code>ControllerPublishVolumeRequest.VolumeAttributes</code></li>
          <li><code>ValidateVolumeCapabilitiesRequest.VolumeAttributes</code></li>
          <li><code>NodePublishVolumeRequest.VolumeAttributes</code></li>
        </ul>
        <p>Enabling this option sets <code>X_CSI_SPEC_VALIDATION=true</code></p>
      </td>
    </tr>
    <tr>
      <td><code>X_CSI_REQUIRE_CREDS</code></td>
      <td>A <code>true</code> value is equivalent to:
        <ul>
          <li><code>X_CSI_REQUIRE_CREDS_CREATE_VOL=true</code></li>
          <li><code>X_CSI_REQUIRE_CREDS_DELETE_VOL=true</code></li>
          <li><code>X_CSI_REQUIRE_CREDS_CTRLR_PUB_VOL=true</code></li>
          <li><code>X_CSI_REQUIRE_CREDS_CTRLR_UNPUB_VOL=true</code></li>
          <li><code>X_CSI_REQUIRE_CREDS_NODE_PUB_VOL=true</code></li>
          <li><code>X_CSI_REQUIRE_CREDS_NODE_UNPUB_VOL=true</code></li>
        </ul>
        <p>Enabling this option sets <code>X_CSI_SPEC_VALIDATION=true</code></p>
      </td>
    </tr>
    <tr>
      <td><code>X_CSI_REQUIRE_CREDS_CREATE_VOL</code></td>
      <td>
        <p>A flag that enables treating the following fields as required:</p>
        <ul><li><code>CreateVolumeRequest.UserCredentials</code></li></ul>
        <p>Enabling this option sets <code>X_CSI_SPEC_VALIDATION=true</code></p>
      </td>
    </tr>
    <tr>
      <td><code>X_CSI_REQUIRE_CREDS_DELETE_VOL</code></td>
      <td>
        <p>A flag that enables treating the following fields as required:</p>
        <ul><li><code>DeleteVolumeRequest.UserCredentials</code></li></ul>
        <p>Enabling this option sets <code>X_CSI_SPEC_VALIDATION=true</code></p>
      </td>
    </tr>
    <tr>
      <td><code>X_CSI_REQUIRE_CREDS_CTRLR_PUB_VOL</code></td>
      <td>
        <p>A flag that enables treating the following fields as required:</p>
        <ul><li><code>ControllerPublishVolumeRequest.UserCredentials</code></li></ul>
        <p>Enabling this option sets <code>X_CSI_SPEC_VALIDATION=true</code></p>
      </td>
    </tr>
    <tr>
      <td><code>X_CSI_REQUIRE_CREDS_CTRLR_UNPUB_VOL</code></td>
      <td>
        <p>A flag that enables treating the following fields as required:</p>
        <ul><li><code>ControllerUnpublishVolumeRequest.UserCredentials</code></li></ul>
        <p>Enabling this option sets <code>X_CSI_SPEC_VALIDATION=true</code></p>
      </td>
    </tr>
    <tr>
      <td><code>X_CSI_REQUIRE_CREDS_NODE_PUB_VOL</code></td>
      <td>
        <p>A flag that enables treating the following fields as required:</p>
        <ul><li><code>NodePublishVolumeRequest.UserCredentials</code></li></ul>
        <p>Enabling this option sets <code>X_CSI_SPEC_VALIDATION=true</code></p>
      </td>
    </tr>
    <tr>
      <td><code>X_CSI_REQUIRE_CREDS_NODE_UNPUB_VOL</code></td>
      <td>
        <p>A flag that enables treating the following fields as required:</p>
        <ul><li><code>NodeUnpublishVolumeRequest.UserCredentials</code></li></ul>
        <p>Enabling this option sets <code>X_CSI_SPEC_VALIDATION=true</code></p>
      </td>
    </tr>
    <tr>
      <td><code>X_CSI_IDEMP</code></td>
      <td>A flag that enables the idempotency interceptor. Even when true,
      the StoragePlugin must still provide a valid IdempotencyProvider
      in order to enable the idempotency interceptor.</td>
    </tr>
    <tr>
      <td><code>X_CSI_IDEMP_TIMEOUT</code></td>
      <td>A <a href="https://golang.org/pkg/time/#ParseDuration"><code>
      time.Duration</code></a> string that determines how long the
      idempotency interceptor waits to obtain a lock for the request's volume
      before returning the gRPC error code <code>FailedPrecondition</code> to
      indicate an operation is already pending for the specified volume.</td>
    </tr>
    <tr>
      <td><code>X_CSI_IDEMP_REQUIRE_VOL</code></td>
      <td>A flag that indicates whether the idempotency interceptor validates
      the existence of a volume before allowing an operation to proceed.</td>
    </tr>
  </tbody>
</table>
