#!/usr/bin/env bash

set -eu -o pipefail

source ./lib.sh

# Runs tests for all combinations of "http" and "https" with and without upstream proxy.
for HTTPBIN_SCHEME in "http" "https" "h2"; do
  for PROXY_SCHEME in "http" "https"; do
    for UPSTREAM_PROXY_SCHEME in "" "http" "https"; do
      run_test "${HTTPBIN_SCHEME}" "${PROXY_SCHEME}" "${UPSTREAM_PROXY_SCHEME}" ""
    done
  done
done

RUN="Localhost" run_test "http" "http" "" "override/proxy-localhost.yaml"
