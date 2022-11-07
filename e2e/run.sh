#!/usr/bin/env bash

set -eu -o pipefail

source ./lib.sh

# Runs tests for all combinations of "http" and "https" with and without upstream proxy.
for HTTPBIN_SCHEME in "http" "https"; do
  for PROXY_SCHEME in "http" "https"; do
    for UPSTREAM_PROXY_SCHEME in "" "http" "https"; do
      run_test "${HTTPBIN_SCHEME}" "${PROXY_SCHEME}" "${UPSTREAM_PROXY_SCHEME}" ""
    done
  done
done

# Runs tests for target or upstream proxy with HTTP/2.
# HTTP/2 implies TLS, so we only test "https" scheme, otherwise it fails with "http2: unsupported scheme".
for HTTPBIN_SCHEME in "https" "h2"; do
  for UPSTREAM_PROXY_SCHEME in "" "h2"; do
    run_test "${HTTPBIN_SCHEME}" "https" "${UPSTREAM_PROXY_SCHEME}" ""
  done
done

RUN="Localhost" run_test "http" "http" "" "override/proxy-localhost.yaml"
