#!/usr/bin/env bash
# Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at https://mozilla.org/MPL/2.0/.

# lib.sh - Common functions for building, packaging, and running a container

set -e -o pipefail

export ROOT_DIR=$(git rev-parse --show-toplevel)

FORCE_BUILD_IMAGE=false
FORCE_RELEASE=false

# Usage: process_flags "$@"
process_flags() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --force-build-image)
                FORCE_BUILD_IMAGE=true
                ;;
            --force-release)
                FORCE_RELEASE=true
                ;;
            *)
                ;;
        esac
        shift
    done
}

# Usage: build_image IMG_NAME DOCKERFILE
function build_image() {
    local img="$1"
    local dockerfile="$2"
    if ! podman image exists ${img} || [ ${FORCE_BUILD_IMAGE} = true ] ; then
        podman build --no-cache -t ${img} -f ${dockerfile}
    fi
}

# Usage: create_package DIST PACKAGE_NAME
function create_package() {
    local dist="$1"
    local package_name="$2"
    if [[ ! -f ${package_name} || ${FORCE_RELEASE} = true ]] ; then
        make -C ${ROOT_DIR} dist
        cp ${dist} ${package_name}
    fi
}

# Usage: run_interactive CONTAINER_NAME
function run_interactive() {
    local container_name="$1"
    podman exec ${container_name} systemctl enable forwarder
    podman exec ${container_name} systemctl start forwarder
    podman exec ${container_name} systemctl status forwarder
    podman exec -it ${container_name} /bin/bash
}

# Usage: remove_container CONTAINER_NAME
function remove_container() {
    local container_name="$1"
    podman rm --force ${container_name}
}
