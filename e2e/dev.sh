#!/usr/bin/env bash

set -eu -o pipefail

source ./lib.sh

HTTPBIN_SCHEME="http"
PROXY_SCHEME="http"

make up CONF="override/httpbin-${HTTPBIN_SCHEME}.yaml override/proxy-${PROXY_SCHEME}.yaml"

docker-compose logs -f