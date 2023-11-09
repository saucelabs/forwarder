#!/bin/sh

set -e

go run ./cmd/forwarder config-file > "forwarder.yaml"
