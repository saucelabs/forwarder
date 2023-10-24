#!/usr/bin/env bash
# Copyright 2023 Sauce Labs Inc. All rights reserved.
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at https://mozilla.org/MPL/2.0/.

set -e -o pipefail

IMG="fedora-systemd:latest"
CONTAINER="forwarder-testing-fedora"
DIST="../../dist/forwarder*_linux.aarch64.rpm"

# Build image if it doesn't exist or FORCE_BUILD_IMAGE is set
if ! podman image exists $IMG || [ -n "$FORCE_BUILD_IMAGE" ]; then
    podman build --no-cache -t $IMG .
fi

# Create forwarder.rpm if it doesn't exist or FORCE_RELEASE is set.
if [[ ! -f forwarder.rpm || -n $FORCE_RELEASE ]]; then
    (cd ../../ && ./bin/goreleaser release --snapshot --skip-docker --clean)
    cp $DIST forwarder.rpm
fi

# Run the container
podman run -p 3128:3128 -d -v ./forwarder.rpm:/forwarder.rpm --name $CONTAINER $IMG
podman exec $CONTAINER dnf -y install /forwarder.rpm
podman exec $CONTAINER systemctl enable forwarder
podman exec $CONTAINER systemctl start forwarder
podman exec $CONTAINER systemctl status forwarder
podman exec -it $CONTAINER /bin/sh
podman rm --force $CONTAINER
