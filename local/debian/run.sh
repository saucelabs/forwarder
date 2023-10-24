#!/usr/bin/env bash
# Copyright 2023 Sauce Labs Inc. All rights reserved.
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at https://mozilla.org/MPL/2.0/.

source ../lib.sh

set -e -o pipefail

IMG="debian-systemd:bullseye"
CONTAINER="forwarder-testing-debian"
DIST="../../dist/forwarder*linux_arm64.deb"

process_flags "$@"

build_image $IMG
create_package $DIST forwarder.deb

podman run -p 3128:3128 -d -v ./forwarder.deb:/forwarder.deb --name $CONTAINER --replace $IMG
podman exec $CONTAINER dpkg -i /forwarder.deb
run_interactive $CONTAINER
podman rm --force $CONTAINER
