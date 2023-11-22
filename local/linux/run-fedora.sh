#!/usr/bin/env bash
# Copyright 2023 Sauce Labs Inc., all rights reserved.
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at https://mozilla.org/MPL/2.0/.

source ./lib.sh

set -e -o pipefail

IMG="fedora-systemd:latest"
CONTAINER="forwarder-testing-fedora"
DIST="$ROOT_DIR/dist/forwarder*_linux.aarch64.rpm"

process_flags "$@"

build_image "$IMG" fedora-systemd.Dockerfile
create_package "$DIST" forwarder.rpm

podman run -p 3128:3128 -d -v ./forwarder.rpm:/forwarder.rpm --name "$CONTAINER" --replace "$IMG"
trap "remove_container $CONTAINER" EXIT

podman exec "$CONTAINER" dnf -y install /forwarder.rpm
run_interactive "$CONTAINER"
