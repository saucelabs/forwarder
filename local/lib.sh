#!/usr/bin/env bash
# Copyright 2023 Sauce Labs Inc. All rights reserved.
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at https://mozilla.org/MPL/2.0/.

# lib.sh - Common functions for building, packaging, and running a container

set -e -o pipefail

force_build_image=false
force_release=false

# Usage: process_flags "$@"
process_flags() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --force-build-image)
                force_build_image=true
                ;;
            --force-release)
                force_release=true
                ;;
            *)
                ;;
        esac
        shift
    done
}

# Usage: build_image IMG_NAME
function build_image() {
    local img="$1"
    if ! podman image exists "$img" || [ "$force_build_image" = true ] ; then
        podman build --no-cache -t "$img" .
    fi
}

# Usage: create_package DIST PACKAGE_NAME
function create_package() {
    local dist="$1"
    local package_name="$2"
    if [[ ! -f "$package_name"  ||  "$force_release" = true ]] ; then
        (cd ../../ && ./bin/goreleaser release --snapshot --skip-docker --clean)
        cp "$dist" "$package_name"
    fi
}

# Usage: run_interactive CONTAINER_NAME
function run_interactive() {
    local container_name="$1"
    podman exec "$container_name" systemctl enable forwarder
    podman exec "$container_name" systemctl start forwarder
    podman exec "$container_name" systemctl status forwarder
    podman exec -it "$container_name" /bin/bash
}
