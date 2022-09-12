# forwarder

The `forwarder` package provides a simple forward proxy.
The proxy can be protected with basic auth.
It can also forward connections to a parent proxy, and authorize connections against that.
Both local and parent credentials can be set via environment variables.
For local proxy credential, set `FORWARDER_LOCALPROXY_AUTH`.
For remote proxy credential, set `FORWARDER_UPSTREAMPROXY_AUTH`.

## Install

### Binary

Download from the [releases page](https://github.com/saucelabs/forwarder/releases).

### Package

`$ go get github.com/saucelabs/forwarder@vX.Y.Z`

```go
import "github.com/saucelabs/forwarder/pkg/proxy"
```

## Usage

See [`example_test.go`](pkg/proxy/example_test.go), and [`proxy_test.go`](pkg/proxy/proxy_test.go) file.

## Documentation

Run `$ make doc` or check out [online](https://pkg.go.dev/github.com/saucelabs/forwarder).

## Development

1. Install required tools by running `make install-dependencies`
2. Run server in development mode `make dev` 

For more information see [CONTRIBUTION](CONTRIBUTION.md).
