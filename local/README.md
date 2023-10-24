# Local Testing Environment
This package provides environments for local testing/developing of the Forwarder.

## Packaging Testing Environment
The environment is based on Docker image with systemd support and is meant to be used with Podman.
It provides a Dockerfile to build the container image and a run script.

### Prerequisites

- [Podman](https://podman.io/) installed on your system

### Supported distributions
- `/debian` - `.deb`
- `/fedora` - `.rpm`

### Getting started
- `./run.sh` will build the container image, release the package and run the container.
- Use `--force-build-image` to always rebuild image before running.
- Use `--force-release` to always release package before running.
