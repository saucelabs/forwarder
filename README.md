# Forwarder Proxy [![Build Status](https://github.com/saucelabs/forwarder/actions/workflows/go.yml/badge.svg)](https://github.com/saucelabs/forwarder/actions/workflows/go.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/saucelabs/forwarder)](https://goreportcard.com/report/github.com/saucelabs/forwarder) [![GitHub release](https://img.shields.io/github/release/saucelabs/forwarder.svg)](https://github.com/saucelabs/forwarder/releases)

Forwarder is a production-ready, fast MITM proxy with PAC support.
It's suitable for debugging, intercepting and manipulating HTTP traffic.
It's used as a core component of Sauce Labs [Sauce Connect Proxy](https://docs.saucelabs.com/secure-connections/sauce-connect/).
It is a forward proxy, which means it proxies traffic from clients to servers (e.g. browsers to websites), and supports `CONNECT` requests.
It can proxy:

* HTTP/HTTPS/HTTP2 requests
* WebSockets (both HTTP and HTTPS)
* Server Sent Events (SSE)
* TCP traffic (e.g. SMTP, IMAP, etc.)

## Documentation

The documentation is available at [forwarder-proxy.io](https://forwarder-proxy.io).

## Development

### Quick Start

- Install Docker and Docker Compose
- Install `make`
- Run `make install-dependencies`

### Linting

- Run `make fmt` to auto format code
- Run `make lint` to lint code
- Edit [.golangci.yml](.golangci.yml) to change linting rules

### Building Devel Images

- Run `make update-devel-image` to build the devel docker image.
  The target supports both `docker` and `podman` as container engines.
  Configure with `CONTAINER_RUNTIME=<docker|podman>`.

### Testing

- Run `make test` to run Go unit tests
- Run `make -C e2e run-e2e` to run e2e tests, more details in [e2e/README.md](e2e/README.md)

### Updating tools versions

All tools versions are defined in [.version](.version) file.
To update a version, edit the file and create a merge request.
CI will run and update the CI image with the new version.
