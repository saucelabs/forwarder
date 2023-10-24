#!/usr/bin/env bash
# Copyright 2023 Sauce Labs Inc. All rights reserved.
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at https://mozilla.org/MPL/2.0/.

set -e -o pipefail

IMG="debian-systemd:bullseye"
CONTAINER="forwarder-testing-debian"
DIST="../../dist/forwarder*linux_arm64.deb"

# Build image if it doesn't exist or FORCE_BUILD_IMAGE is set
if ! podman image exists $IMG || [ -n "$FORCE_BUILD_IMAGE" ]; then
    podman build --no-cache -t $IMG .
fi

# Create forwarder.deb if it doesn't exist or FORCE_RELEASE is set.
if [[ ! -f forwarder.deb || -n $FORCE_RELEASE ]]; then
    (cd ../../ && ./bin/goreleaser release --snapshot --skip-docker --clean)
    cp $DIST forwarder.deb
fi

# Run the container
podman run -p 3128:3128 -d -v ./forwarder.deb:/forwarder.deb --name $CONTAINER $IMG
podman exec $CONTAINER dpkg -i /forwarder.deb
podman exec $CONTAINER systemctl enable forwarder
podman exec $CONTAINER systemctl start forwarder
podman exec $CONTAINER systemctl status forwarder
podman exec -it $CONTAINER /bin/sh
podman rm --force $CONTAINER
