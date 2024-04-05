# Linux Installation Environment

This directory contains the development environment for the Linux installation of Forwarder.
It's based on Systemd enabled Podman containers.
Podman installation is required to use this environment.

## Supported distributions
- Fedora `.rpm`
- Debian `.deb`

## Usage

1. Create packages for the distributions you want to test with `make dist`.
1. Run the containers with `make debian` or `make fedora`, this will create the containers install packages and start the services.
1. Check the Makefile for available commands. Example commands include:
	- `make shell` to enter the container
	- `make logs` to see the logs of the service
1. To stop the containers run `make down`.