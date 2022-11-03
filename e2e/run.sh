#!/usr/bin/env bash

set -eu -o pipefail

on_error() {
  make dump-logs
}

run_direct_test() {
  trap 'on_error' ERR

  PROXY_SCHEME=$1
  HTTPBIN_SCHEME=$1

  echo ">>> DIRECT proxy=${PROXY_SCHEME} httpbin=${HTTPBIN_SCHEME}"
  make up CONF="override/proxy-${PROXY_SCHEME}.yaml httpbin-${HTTPBIN_SCHEME}.yaml"
  make test ARGS="-proxy ${PROXY_SCHEME}://proxy:3128 -httpbin ${HTTPBIN_SCHEME}://httpbin -insecure-skip-verify"
}

run_direct_test http http
run_direct_test https http
run_direct_test https https
run_direct_test http https
