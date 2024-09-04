#!/bin/sh

set -e

go run -ldflags "-checklinkname=0" ./cmd/forwarder run config-file > "forwarder.yaml"
