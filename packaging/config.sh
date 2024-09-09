#!/bin/sh

set -e

go run ./cmd/forwarder run config-file > "forwarder.yaml"
