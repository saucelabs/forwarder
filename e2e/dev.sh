#!/usr/bin/env bash

set -eu -o pipefail

source ./lib.sh

HTTPBIN_SCHEME="https"
PROXY_SCHEME="https"

make up CONF="httpbin-${HTTPBIN_SCHEME}.yaml override/proxy-${PROXY_SCHEME}.yaml"

docker-compose logs -f